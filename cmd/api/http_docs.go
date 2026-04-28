package main

import (
	"context"
	"net/http"

	"fraud-detection/internal/messaging/rabbitmq"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type healthUpResponse struct {
	Postgres string `json:"postgres" example:"up" extensions:"x-order=0"`
	Redis    string `json:"redis" example:"up" extensions:"x-order=1"`
	RabbitMQ string `json:"rabbitmq" example:"up" extensions:"x-order=2"`
}

type healthDownResponse struct {
	Postgres string `json:"postgres" example:"down" extensions:"x-order=0"`
	Redis    string `json:"redis" example:"down" extensions:"x-order=1"`
	RabbitMQ string `json:"rabbitmq" example:"down" extensions:"x-order=2"`
}

// healthHandler godoc
// @Summary Check API health
// @Description Checks PostgreSQL, Redis, and RabbitMQ connectivity.
// @Tags system
// @Produce json
// @Success 200 {object} healthUpResponse
// @Failure 503 {object} healthDownResponse
// @Router /health [get]
func healthHandler(db *gorm.DB, redisClient *redis.Client, rabbitMQClient *rabbitmq.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := healthDownResponse{
			Postgres: "down",
			Redis:    "down",
			RabbitMQ: "down",
		}

		sqlDB, err := db.DB()
		if err == nil && sqlDB.Ping() == nil {
			status.Postgres = "up"
		}

		if err := redisClient.Ping(context.Background()).Err(); err == nil {
			status.Redis = "up"
		}

		if err := rabbitMQClient.Ping(); err == nil {
			status.RabbitMQ = "up"
		}

		if status.Postgres == "up" && status.Redis == "up" && status.RabbitMQ == "up" {
			c.JSON(http.StatusOK, healthUpResponse{
				Postgres: status.Postgres,
				Redis:    status.Redis,
				RabbitMQ: status.RabbitMQ,
			})
			return
		}

		c.JSON(http.StatusServiceUnavailable, status)
	}
}
