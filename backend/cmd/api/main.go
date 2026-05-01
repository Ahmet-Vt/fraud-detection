package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	_ "fraud-detection/docs"
	"fraud-detection/internal/cache"
	"fraud-detection/internal/config"
	"fraud-detection/internal/database"
	"fraud-detection/internal/fraud"
	"fraud-detection/internal/messaging/rabbitmq"
	"fraud-detection/internal/transaction"
	"fraud-detection/internal/user"
	"fraud-detection/internal/worker"
	"fraud-detection/migrations"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// @title Fraud Detection API
// @version 1.0
// @description HTTP API for the fraud detection backend.
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	cfg := config.Load()

	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := runMigrations(db); err != nil {
		log.Fatal(err)
	}

	redisClient, err := initRedis(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer closeRedis(redisClient)

	rabbitMQClient, err := initRabbitMQ(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer closeRabbitMQ(rabbitMQClient)

	publisher := rabbitmq.NewPublisher(rabbitMQClient)
	defer closePublisher(publisher)

	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	transactionRepository := transaction.NewRepository(db)
	transactionStateCache := cache.NewTransactionStateCache(redisClient)
	transactionService := transaction.NewService(transactionRepository, publisher, transactionStateCache)
	transactionHandler := transaction.NewHandler(transactionService)
	fraudService := fraud.NewService(transactionService, transactionStateCache)
	workerPool := worker.NewFraudWorker(fraudService, transactionService)

	startFraudConsumers(cfg, rabbitMQClient, workerPool)

	router := gin.Default()
	router.GET("/health", healthHandler(db, redisClient, rabbitMQClient))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	userHandler.RegisterRoutes(router)
	transactionHandler.RegisterRoutes(router)

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}

func initDatabase(cfg config.AppConfig) (*gorm.DB, error) {
	db, err := database.NewPostgres(cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql database instance: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.DB.ConnMaxIdleTime) * time.Minute)

	return db, nil
}

func runMigrations(db *gorm.DB) error {
	migrator := database.NewMigrator(db)
	if err := migrations.Register(migrator); err != nil {
		return fmt.Errorf("register migrations: %w", err)
	}

	if err := migrator.Run(); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func initRedis(cfg config.AppConfig) (*redis.Client, error) {
	redisClient, err := cache.NewRedis(
		context.Background(),
		cfg.Redis.Addr(),
	)

	if err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	return redisClient, nil
}

func closeRedis(redisClient *redis.Client) {
	if err := redisClient.Close(); err != nil {
		log.Printf("failed to close redis client: %v", err)
	}
}

func initRabbitMQ(cfg config.AppConfig) (*rabbitmq.Client, error) {
	rabbitMQClient, err := rabbitmq.New(cfg.RabbitMQ.AMQPURL())
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	if err := rabbitMQClient.DeclareTopology(); err != nil {
		return nil, fmt.Errorf("declare rabbitmq topology: %w", err)
	}

	return rabbitMQClient, nil
}

func closeRabbitMQ(rabbitMQClient *rabbitmq.Client) {
	if err := rabbitMQClient.Close(); err != nil {
		log.Printf("failed to close rabbitmq client: %v", err)
	}
}

func closePublisher(publisher *rabbitmq.Publisher) {
	if err := publisher.Close(); err != nil {
		log.Printf("failed to close rabbitmq publisher: %v", err)
	}
}

func startFraudConsumers(cfg config.AppConfig, rabbitMQClient *rabbitmq.Client, workerPool *worker.FraudWorker) {
	for i := 1; i <= cfg.RabbitMQ.ConsumerCount; i++ {
		tag := fmt.Sprintf("fraud-worker-%d", i)
		fraudConsumer := rabbitmq.NewConsumer(
			rabbitMQClient,
			func(ctx context.Context, evt rabbitmq.TransactionCreatedMessage) error {
				return workerPool.HandleTransactionCreated(ctx, evt)
			},
			tag,
			cfg.RabbitMQ.PrefetchCount,
		)

		go func(consumerTag string, consumer *rabbitmq.Consumer) {
			if err := consumer.Run(context.Background()); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("rabbitmq consumer stopped tag=%s: %v", consumerTag, err)
			}
		}(tag, fraudConsumer)
	}
}
