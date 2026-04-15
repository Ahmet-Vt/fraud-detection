package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type pingResponse struct {
	Message string `json:"message" example:"pong" extensions:"x-order=0"`
}

type healthUpResponse struct {
	Postgres string `json:"postgres" example:"up" extensions:"x-order=0"`
}

type healthDownResponse struct {
	Postgres string `json:"postgres" example:"down" extensions:"x-order=0"`
}

// pingHandler godoc
// @Summary Ping the API
// @Description Returns a simple pong response.
// @Tags system
// @Produce json
// @Success 200 {object} pingResponse
// @Router /ping [get]
func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, pingResponse{Message: "pong"})
}

// healthHandler godoc
// @Summary Check API health
// @Description Checks PostgreSQL connectivity.
// @Tags system
// @Produce json
// @Success 200 {object} healthUpResponse
// @Failure 503 {object} healthDownResponse
// @Router /health [get]
func healthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, healthDownResponse{Postgres: "down"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, healthDownResponse{Postgres: "down"})
			return
		}

		c.JSON(http.StatusOK, healthUpResponse{Postgres: "up"})
	}
}
