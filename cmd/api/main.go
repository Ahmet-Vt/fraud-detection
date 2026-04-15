package main

import (
	"log"

	_ "github.com/Ahmet-Vt/fraud-detection/docs"
	"github.com/Ahmet-Vt/fraud-detection/internal/config"
	"github.com/Ahmet-Vt/fraud-detection/internal/database"
	"github.com/Ahmet-Vt/fraud-detection/internal/transaction"
	"github.com/Ahmet-Vt/fraud-detection/internal/user"
	"github.com/Ahmet-Vt/fraud-detection/migrations"
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

	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)


	transactionRepository := transaction.NewRepository(db)
	transactionService := transaction.NewService(transactionRepository)
	transactionHandler := transaction.NewHandler(transactionService)

	router := gin.Default()

	router.GET("/ping", pingHandler)
	router.GET("/health", healthHandler(db))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userHandler.RegisterRoutes(router)
	transactionHandler.RegisterRoutes(router)

	if err := router.Run(); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
