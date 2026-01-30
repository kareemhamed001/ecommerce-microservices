package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/e-commerce/pkg/db"
	"github.com/kareemhamed001/e-commerce/pkg/jwt"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/pkg/tracer"
	"github.com/kareemhamed001/e-commerce/services/UserService/config"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/delivery/grpc/handler"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/domain"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/repository/postgresql"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/usecase"
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
		MigrationDir:          "services/UserService/internal/migrations",
		ConnectionMaxIdle:     config.DBConnectionMaxIdle,
		ConnectionMaxOpen:     config.DBConnectionMaxOpen,
		ConnectionMaxLifeTime: config.DBConnectionMaxLife,
	}

	db, err := db.InitDB(dbConfig)
	if err != nil {
		close(done)
		panic("failed to connect database")
	}

	db.AutoMigrate(&domain.User{}, &domain.Address{})

	useRepo := postgresql.NewUserRepository(db)
	addressRepo := postgresql.NewAddressRepository(db)
	userUseCase := usecase.NewUserUsecase(useRepo)
	addressUsecase := usecase.NewAddressUsecase(addressRepo)

	validate := validator.New()
	jwtManager := jwt.NewJWTManager(config.JWTSecret, time.Duration(config.JWTDuration)*time.Hour)

	grpcHandler := handler.NewUserGRPCHandler(userUseCase, addressUsecase, validate, jwtManager)

	err = grpcHandler.Run(done, config.GRPCPort)
	if err != nil {
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
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "jaeger:4317")
	tp, err := tracer.InitTracer(ctx, "user-service-grpc", jaegerEndpoint)
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
