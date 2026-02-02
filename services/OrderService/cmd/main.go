package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/e-commerce/pkg/db"
	"github.com/kareemhamed001/e-commerce/pkg/grpcmiddleware"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/pkg/tracer"
	"github.com/kareemhamed001/e-commerce/services/OrderService/config"
	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/delivery/grpc/handler"
	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/domain"
	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/repository/postgresql"
	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/usecase"
	productpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/product"
	userpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	done := make(chan interface{})
	config, err := config.Load()
	if err != nil {
		close(done)
		panic(err)
	}

	logger.InitGlobal(config.AppEnv, "logs/order/system.log")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracer := initTracing(ctx)
	defer shutdownTracer()

	dbConfig := &db.Config{
		DBDriver:              config.DBDriver,
		DSN:                   config.DBDSN,
		MigrationAutoRun:      config.DBMigrationAutoRun,
		MigrationDir:          "services/OrderService/internal/migrations",
		ConnectionMaxIdle:     config.DBConnectionMaxIdle,
		ConnectionMaxOpen:     config.DBConnectionMaxOpen,
		ConnectionMaxLifeTime: config.DBConnectionMaxLife,
	}

	orderDB, err := db.InitDB(dbConfig)
	if err != nil {
		close(done)
		panic("failed to connect database")
	}

	orderDB.AutoMigrate(&domain.Order{}, &domain.OrderItem{})

	productConn, err := grpc.NewClient(
		config.ProductServiceGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcmiddleware.InternalAuthUnaryClientInterceptor(config.InternalAuthToken),
			grpcmiddleware.CircuitBreakerUnaryClientInterceptor(
				"order-service->"+config.ProductServiceGRPCAddr,
				grpcmiddleware.CircuitBreakerConfig{
					Enabled:      config.CircuitBreakerEnabled,
					MaxRequests:  config.CircuitBreakerMaxRequests,
					Interval:     config.CircuitBreakerInterval,
					Timeout:      config.CircuitBreakerTimeout,
					FailureRatio: config.CircuitBreakerFailureRatio,
					MinRequests:  config.CircuitBreakerMinRequests,
				},
			),
		),
	)
	if err != nil {
		close(done)
		panic("failed to connect to product service")
	}
	defer func() {
		_ = productConn.Close()
	}()

	userConn, err := grpc.Dial(
		config.UserServiceGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcmiddleware.InternalAuthUnaryClientInterceptor(config.InternalAuthToken),
			grpcmiddleware.CircuitBreakerUnaryClientInterceptor(
				"order-service->"+config.UserServiceGRPCAddr,
				grpcmiddleware.CircuitBreakerConfig{
					Enabled:      config.CircuitBreakerEnabled,
					MaxRequests:  config.CircuitBreakerMaxRequests,
					Interval:     config.CircuitBreakerInterval,
					Timeout:      config.CircuitBreakerTimeout,
					FailureRatio: config.CircuitBreakerFailureRatio,
					MinRequests:  config.CircuitBreakerMinRequests,
				},
			),
		),
	)
	if err != nil {
		close(done)
		panic("failed to connect to user service")
	}
	defer func() {
		_ = userConn.Close()
	}()

	orderRepo := postgresql.NewOrderRepository(orderDB)
	productClient := productpb.NewProductServiceClient(productConn)
	userClient := userpb.NewUserServiceClient(userConn)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, productClient, userClient)

	validate := validator.New()
	grpcHandler := handler.NewOrderGRPCHandler(orderUsecase, validate, config.InternalAuthToken)

	if err := grpcHandler.Run(done, config.GRPCPort); err != nil {
		logger.Errorf("failed to start gRPC server: %v", err)
		close(done)
		panic(err)
	}

	// graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	close(done)
	time.Sleep(200 * time.Millisecond)
}

func initTracing(ctx context.Context) func() {
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "ecommece_jaeger:4317")
	tp, err := tracer.InitTracer(ctx, "order-service-grpc", jaegerEndpoint)
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
