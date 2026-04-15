package main

import (
	"log"
	"net/http"

	"github.com/Ahmet-Vt/fraud-detection/internal/config"
	"github.com/Ahmet-Vt/fraud-detection/internal/database"
	"github.com/Ahmet-Vt/fraud-detection/internal/user"
	"github.com/Ahmet-Vt/fraud-detection/migrations"
	"github.com/gin-gonic/gin"
)

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

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"postgres": "error: " + err.Error()})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"postgres": "down: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"postgres": "up"})
	})

	userHandler.RegisterRoutes(router)

	if err := router.Run(); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
