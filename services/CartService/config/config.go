package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
)

type Config struct {
	// Server
	AppPort string
	AppEnv  string

	// Redis
	RedisEnabled  bool
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// gRPC
	GRPCPort string

	// Downstream gRPC services
	ProductServiceGRPCAddr string
	UserServiceGRPCAddr    string

	// Service name
	ServiceName string

	// Internal service auth
	InternalAuthToken string

	// Timeouts
	DownstreamTimeout time.Duration

	// Circuit breaker
	CircuitBreakerEnabled      bool
	CircuitBreakerMaxRequests  uint32
	CircuitBreakerInterval     time.Duration
	CircuitBreakerTimeout      time.Duration
	CircuitBreakerFailureRatio float64
	CircuitBreakerMinRequests  uint32
}

func Load() (*Config, error) {
	envPaths := []string{
		filepath.Join("services/CartService/config/.env"),
		filepath.Join("config/.env"),
		filepath.Join("./.env"),
	}

	var err error
	for _, envPath := range envPaths {
		err = godotenv.Load(envPath)
		if err == nil {
			logger.Infof("loaded .env file from: %s", envPath)
			break
		}
	}

	if err != nil {
		logger.Warnf("could not load .env file from any path: %v", err)
	}

	cfg := &Config{
		AppPort: GetEnv("APP_PORT", "8086"),
		AppEnv:  GetEnv("APP_ENV", "development"),

		RedisEnabled:  getEnvBool("REDIS_ENABLED", true),
		RedisHost:     GetEnv("REDIS_HOST", "localhost"),
		RedisPort:     GetEnv("REDIS_PORT", "6379"),
		RedisPassword: GetEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		GRPCPort: GetEnv("GRPC_PORT", "50057"),

		ProductServiceGRPCAddr: GetEnv("PRODUCT_SERVICE_GRPC_ADDR", "localhost:50053"),
		UserServiceGRPCAddr:    GetEnv("USER_SERVICE_GRPC_ADDR", "localhost:50051"),

		ServiceName:       GetEnv("SERVICE_NAME", "cart-service"),
		DownstreamTimeout: time.Duration(getEnvInt("DOWNSTREAM_TIMEOUT_SECONDS", 3)) * time.Second,

		InternalAuthToken: GetEnv("INTERNAL_AUTH_TOKEN", ""),

		CircuitBreakerEnabled:      getEnvBool("CB_ENABLED", true),
		CircuitBreakerMaxRequests:  uint32(getEnvInt("CB_MAX_REQUESTS", 5)),
		CircuitBreakerInterval:     time.Duration(getEnvInt("CB_INTERVAL_SECONDS", 60)) * time.Second,
		CircuitBreakerTimeout:      time.Duration(getEnvInt("CB_TIMEOUT_SECONDS", 20)) * time.Second,
		CircuitBreakerFailureRatio: getEnvFloat("CB_FAILURE_RATIO", 0.6),
		CircuitBreakerMinRequests:  uint32(getEnvInt("CB_MIN_REQUESTS", 20)),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.GRPCPort == "" {
		return fmt.Errorf("GRPC_PORT is required")
	}

	if c.AppPort == "" {
		return fmt.Errorf("APP_PORT is required")
	}

	if c.ProductServiceGRPCAddr == "" {
		return fmt.Errorf("PRODUCT_SERVICE_GRPC_ADDR is required")
	}

	if c.UserServiceGRPCAddr == "" {
		return fmt.Errorf("USER_SERVICE_GRPC_ADDR is required")
	}

	if c.RedisHost == "" || c.RedisPort == "" {
		return fmt.Errorf("REDIS_HOST and REDIS_PORT are required")
	}

	if c.InternalAuthToken == "" {
		return fmt.Errorf("INTERNAL_AUTH_TOKEN is required")
	}

	return nil
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		var intVal int
		_, err := fmt.Sscanf(value, "%d", &intVal)
		if err != nil {
			return fallback
		}
		return intVal
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		return value == "true" || value == "1" || value == "yes"
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if value, ok := os.LookupEnv(key); ok {
		var floatVal float64
		_, err := fmt.Sscanf(value, "%f", &floatVal)
		if err != nil {
			return fallback
		}
		return floatVal
	}
	return fallback
}
