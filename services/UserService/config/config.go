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

	// Database
	DBDriver            string
	DBDSN               string
	DBConnectionMaxIdle int
	DBConnectionMaxOpen int
	DBConnectionMaxLife time.Duration
	DBMigrationAutoRun  bool

	// JWT
	JWTSecret   string
	JWTDuration int

	// gRPC
	GRPCPort string

	// Service name
	ServiceName string
}

func Load() (*Config, error) {
	// Try multiple paths for .env file
	envPaths := []string{
		filepath.Join("services/UserService/config/.env"),
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
		// Server
		AppPort: GetEnv("APP_PORT", "8080"),
		AppEnv:  GetEnv("APP_ENV", "development"),

		// Database
		DBDriver:            GetEnv("DB_DRIVER", "postgres"),
		DBDSN:               GetEnv("DB_DSN", "host=localhost user=postgres password=postgres dbname=userservice port=5432 sslmode=disable TimeZone=UTC"),
		DBConnectionMaxIdle: getEnvInt("DB_MAX_IDLE_CONNS", 10),
		DBConnectionMaxOpen: getEnvInt("DB_MAX_OPEN_CONNS", 100),
		DBConnectionMaxLife: time.Duration(getEnvInt("DB_MAX_CONN_LIFETIME_MINUTES", 60)) * time.Minute,
		DBMigrationAutoRun:  getEnvBool("DB_MIGRATION_AUTO_RUN", true),

		// JWT
		JWTSecret:   GetEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTDuration: getEnvInt("JWT_DURATION_HOURS", 24),

		// gRPC
		GRPCPort: GetEnv("GRPC_PORT", "50051"),

		// Service
		ServiceName: GetEnv("SERVICE_NAME", "user-service"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DBDriver == "" {
		return fmt.Errorf("DB_DRIVER is required")
	}

	if c.DBDSN == "" {
		return fmt.Errorf("DB_DSN is required")
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if c.AppPort == "" {
		return fmt.Errorf("APP_PORT is required")
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
