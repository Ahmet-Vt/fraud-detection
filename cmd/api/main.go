package main

import (
	"context"
	"errors"
	"log"

	_ "fraud-detection/docs"
	"fraud-detection/internal/config"
	"fraud-detection/internal/database"
	"fraud-detection/internal/messaging/rabbitmq"
	"fraud-detection/internal/transaction"
	"fraud-detection/internal/user"
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

	migrator := database.NewMigrator(db)
	if err := migrations.Register(migrator); err != nil {
		log.Fatal("failed to register migrations: ", err)
	}

	if err := migrator.Run(); err != nil {
		log.Fatal("migrations failed: ", err)
	}

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

	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	transactionRepository := transaction.NewRepository(db)
	transactionService := transaction.NewService(transactionRepository, publisher)
	transactionHandler := transaction.NewHandler(transactionService)

	fraudConsumer := rabbitmq.NewConsumer(rabbitMQClient, func(ctx context.Context, evt rabbitmq.TransactionCreatedMessage) error {
		log.Printf("transaction.created id=%s user=%s amount=%d loc=%s",
			evt.TransactionID, evt.UserID, evt.Amount, evt.Location)
		return nil
	})
	go func() {
		if err := fraudConsumer.Run(context.Background()); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("rabbitmq consumer stopped: %v", err)
		}
	}()

	router := gin.Default()

	router.GET("/health", healthHandler(db, rabbitMQClient))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userHandler.RegisterRoutes(router)
	transactionHandler.RegisterRoutes(router)

	if err := router.Run(); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
