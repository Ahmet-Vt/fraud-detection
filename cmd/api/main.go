package main

import (
	"log"
	"net/http"

	"github.com/Ahmet-Vt/fraud-detection/database"
	"github.com/Ahmet-Vt/fraud-detection/config"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	cfg := config.LoadConfig()

	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

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

	router.Run()
}
