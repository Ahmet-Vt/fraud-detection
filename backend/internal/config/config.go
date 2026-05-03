package config

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}

type RabbitMQConfig struct {
	Host          string
	Port          string
	User          string
	Password      string
	ConsumerCount int
	PrefetchCount int
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
	DB              DBConfig
	RabbitMQ        RabbitMQConfig
	Redis           RedisConfig
	HTTPListenAddr  string
	MCPStdioEnabled bool
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
			DBHost:          requireEnv("POSTGRES_HOST"),
			DBPort:          requireEnv("POSTGRES_PORT"),
			DBUser:          requireEnv("POSTGRES_USER"),
			DBPassword:      requireEnv("POSTGRES_PASSWORD"),
			DBName:          requireEnv("POSTGRES_DB"),
			DBSSLMode:       requireEnv("POSTGRES_SSLMODE"),
			MaxOpenConns:    envIntOrDefault("POSTGRES_MAX_OPEN_CONNS", 40),
			MaxIdleConns:    envIntOrDefault("POSTGRES_MAX_IDLE_CONNS", 20),
			ConnMaxLifetime: envIntOrDefault("POSTGRES_CONN_MAX_LIFETIME_MINUTES", 30),
			ConnMaxIdleTime: envIntOrDefault("POSTGRES_CONN_MAX_IDLE_TIME_MINUTES", 5),
		},
		RabbitMQ: RabbitMQConfig{
			Host:          requireEnv("RABBITMQ_HOST"),
			Port:          requireEnv("RABBITMQ_PORT"),
			User:          requireEnv("RABBITMQ_USER"),
			Password:      requireEnv("RABBITMQ_PASSWORD"),
			ConsumerCount: envIntNonNegativeOrDefault("RABBITMQ_CONSUMER_COUNT", 4),
			PrefetchCount: envIntOrDefault("RABBITMQ_PREFETCH_COUNT", 10),
		},
		Redis: RedisConfig{
			Host: requireEnv("REDIS_HOST"),
			Port: requireEnv("REDIS_PORT"),
		},
		HTTPListenAddr:  envStringOrDefault("HTTP_LISTEN", ":8080"),
		MCPStdioEnabled: envBoolOrDefault("MCP_STDIO_ENABLED", false),
	}
}

func envStringOrDefault(key, defaultValue string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultValue
	}
	return val
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return val
}

func envBoolOrDefault(key string, defaultValue bool) bool {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultValue
	}
	switch strings.ToLower(val) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		log.Printf("invalid %s=%q, using default %v", key, val, defaultValue)
		return defaultValue
	}
}

func envIntOrDefault(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		log.Printf("invalid %s=%q, using default %d", key, val, defaultValue)
		return defaultValue
	}

	return parsed
}

func envIntNonNegativeOrDefault(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed < 0 {
		log.Printf("invalid %s=%q, using default %d", key, val, defaultValue)
		return defaultValue
	}
	return parsed
}
