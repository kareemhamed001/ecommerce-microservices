package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// go:embed services/*/internal/migrations/*.sql
var embedMigrations embed.FS

type Config struct {
	DBDriver              string
	DSN                   string
	MigrationAutoRun      bool
	MigrationDir          string
	ConnectionMaxIdle     int
	ConnectionMaxOpen     int
	ConnectionMaxLifeTime time.Duration
}

// NewDefaultConfig returns a database configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		DBDriver:              "postgres",
		DSN:                   "host=db user=postgres password=postgres dbname=userservice port=5432 sslmode=disable TimeZone=UTC",
		MigrationAutoRun:      true,
		MigrationDir:          "services/UserService/internal/migrations",
		ConnectionMaxIdle:     10,
		ConnectionMaxOpen:     100,
		ConnectionMaxLifeTime: time.Hour,
	}
}

// InitDB initializes the database connection with the provided configuration
func InitDB(cfg *Config) (*gorm.DB, error) {
	gormLogger := logger.NewGormLoggerFromGlobal().
		SetLogLevel(gormlogger.Info).
		SetSlowThreshold(200 * time.Millisecond).
		SetIgnoreRecordNotFoundError(true)
	logger.Infof("connecting to database with DSN: %s", cfg.DSN)
	// Open the database connection
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get the underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	if err := configureConnectionPool(sqlDB, cfg); err != nil {
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}
	// Run migrations if auto-run is enabled
	if cfg.MigrationAutoRun {
		if err := runMigrations(sqlDB, cfg.MigrationDir); err != nil {
			logger.Errorf("failed to run migrations: %v", err)
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	logger.Info("Database connected successfully")
	return db, nil
}

// configureConnectionPool sets up the database connection pool parameters
func configureConnectionPool(db *sql.DB, cfg *Config) error {
	db.SetMaxIdleConns(cfg.ConnectionMaxIdle)
	db.SetMaxOpenConns(cfg.ConnectionMaxOpen)
	db.SetConnMaxLifetime(cfg.ConnectionMaxLifeTime)

	// Verify the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// runMigrations executes the database migrations using goose
func runMigrations(db *sql.DB, migrationDir string) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Errorf("failed to set goose dialect: %v", err)
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if migrationDir == "" {
		migrationDir = "services/UserService/internal/migrations"
	}

	// goose needs the migrations in the current working directory structure
	// Since we embedded with the full path, we use that path
	if err := goose.Up(db, migrationDir); err != nil {
		logger.Warnf("migration warning: %v", err)
		// Don't fail on migration errors in development
	}

	logger.Info("Migrations processed")
	return nil
}

// CloseDB gracefully closes the database connection
func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	log.Println("Database connection closed")
	return nil
}
