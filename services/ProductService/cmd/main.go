package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/e-commerce/pkg/db"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/pkg/redis"
	"github.com/kareemhamed001/e-commerce/pkg/tracer"
	"github.com/kareemhamed001/e-commerce/services/ProductService/config"
	redisCache "github.com/kareemhamed001/e-commerce/services/ProductService/internal/cache/redis"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/delivery/grpc/handler"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/domain"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/repository/postgresql"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/usecase"
)

func main() {
	done := make(chan interface{})
	config, err := config.Load()
	if err != nil {
		close(done)
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracer := initTracing(ctx)
	defer shutdownTracer()

	dbConfig := &db.Config{
		DBDriver:              config.DBDriver,
		DSN:                   config.DBDSN,
		MigrationAutoRun:      config.DBMigrationAutoRun,
		MigrationDir:          "services/ProductService/internal/migrations",
		ConnectionMaxIdle:     config.DBConnectionMaxIdle,
		ConnectionMaxOpen:     config.DBConnectionMaxOpen,
		ConnectionMaxLifeTime: config.DBConnectionMaxLife,
	}

	db, err := db.InitDB(dbConfig)
	if err != nil {
		close(done)
		panic("failed to connect database")
	}

	db.AutoMigrate(&domain.Product{})

	productRepo := postgresql.NewProductRepository(db)
	redisClient, err := redis.NewClient(config)

	if err != nil {
		close(done)
		panic("failed to connect to redis")
	}

	productCache := redisCache.NewProductCache(redisClient)
	productUseCase := usecase.NewProductUsecase(productRepo, productCache)

	categoryRepo := postgresql.NewCategoryRepository(db)
	categoryUseCase := usecase.NewCategoryUsecase(categoryRepo)

	validate := validator.New()

	grpcHandler := handler.NewProductGRPCHandler(productUseCase, categoryUseCase, validate)

	err = grpcHandler.Run(done, config.GRPCPort)
	if err != nil {
		logger.Errorf("failed to start gRPC server: %v", err)
		close(done)
		panic(err)
	}

	//gracful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	close(done)

}

func initTracing(ctx context.Context) func() {
	// For OTLP gRPC, endpoint should be just host:port without http:// scheme or path
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "ecommece_jaeger:4317")
	tp, err := tracer.InitTracer(ctx, "product-service-grpc", jaegerEndpoint)
	if err != nil {
		logger.Warnf("Failed to initialize tracer: %v. Continuing without tracing.", err)
		return func() {}
	}

	logger.Info("OpenTelemetry tracer initialized successfully")
	return func() {
		if err := tracer.Shutdown(ctx, tp); err != nil {
			logger.Errorf("Failed to shutdown tracer: %v", err)
		}
	}
}
