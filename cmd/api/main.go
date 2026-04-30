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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Fraud Detection API
// @version 1.0
// @description HTTP API for the fraud detection backend.
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(cfg.DB.DSN())
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database sql instance: ", err)
	}
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.DB.ConnMaxIdleTime) * time.Minute)

	migrator := database.NewMigrator(db)
	if err := migrations.Register(migrator); err != nil {
		log.Fatal("failed to register migrations: ", err)
	}

	if err := migrator.Run(); err != nil {
		log.Fatal("migrations failed: ", err)
	}

	redisClient, err := cache.NewRedis(
		context.Background(),
		cfg.Redis.Addr(),
	)

	if err != nil {
		log.Fatal("failed to connect redis: ", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("failed to close redis client: %v", err)
		}
	}()

	rabbitMQClient, err := rabbitmq.New(cfg.RabbitMQ.AMQPURL())
	if err != nil {
		log.Fatal("failed to connect rabbitmq: ", err)
	}
	defer func() {
		if err := rabbitMQClient.Close(); err != nil {
			log.Printf("failed to close rabbitmq client: %v", err)
		}
	}()

	if err := rabbitMQClient.DeclareTopology(); err != nil {
		log.Fatal("failed to declare rabbitmq topology: ", err)
	}

	publisher := rabbitmq.NewPublisher(rabbitMQClient)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("failed to close rabbitmq publisher: %v", err)
		}
	}()

	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	transactionRepository := transaction.NewRepository(db)
	transactionStateCache := cache.NewTransactionStateCache(redisClient)
	transactionService := transaction.NewService(transactionRepository, publisher, transactionStateCache)
	transactionHandler := transaction.NewHandler(transactionService)
	fraudService := fraud.NewService(transactionService, transactionStateCache)
	workerPool := worker.NewPool(fraudService)

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

	router := gin.Default()

	router.GET("/health", healthHandler(db, redisClient, rabbitMQClient))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userHandler.RegisterRoutes(router)
	transactionHandler.RegisterRoutes(router)

	if err := router.Run(); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
