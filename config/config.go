package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

type AppConfig struct {
	DB DBConfig
}

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}
}

func LoadConfig() AppConfig {
	return AppConfig{
		DB: DBConfig{
			DBHost:     requireEnv("POSTGRES_HOST"),
			DBPort:     requireEnv("POSTGRES_PORT"),
			DBUser:     requireEnv("POSTGRES_USER"),
			DBPassword: requireEnv("POSTGRES_PASSWORD"),
			DBName:     requireEnv("POSTGRES_DB"),
			DBSSLMode:  requireEnv("POSTGRES_SSLMODE"),
		},
	}
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return val
}
