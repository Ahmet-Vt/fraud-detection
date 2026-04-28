package config

import (
	"fmt"
	"log"
	"net"
	"net/url"
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

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

type RedisConfig struct {
	Host string
	Port string
}

func (c RedisConfig) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c RabbitMQConfig) AMQPURL() string {
	return (&url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, c.Port),
		Path:   "/",
	}).String()
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

type AppConfig struct {
	DB       DBConfig
	RabbitMQ RabbitMQConfig
	Redis    RedisConfig
}

func Load() AppConfig {
	loadEnv()
	return loadConfig()
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}
}

func loadConfig() AppConfig {
	return AppConfig{
		DB: DBConfig{
			DBHost:     requireEnv("POSTGRES_HOST"),
			DBPort:     requireEnv("POSTGRES_PORT"),
			DBUser:     requireEnv("POSTGRES_USER"),
			DBPassword: requireEnv("POSTGRES_PASSWORD"),
			DBName:     requireEnv("POSTGRES_DB"),
			DBSSLMode:  requireEnv("POSTGRES_SSLMODE"),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     requireEnv("RABBITMQ_HOST"),
			Port:     requireEnv("RABBITMQ_PORT"),
			User:     requireEnv("RABBITMQ_USER"),
			Password: requireEnv("RABBITMQ_PASSWORD"),
		},
		Redis: RedisConfig{
			Host: requireEnv("REDIS_HOST"),
			Port: requireEnv("REDIS_PORT"),
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
