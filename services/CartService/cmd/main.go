package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	redisClient "github.com/kareemhamed001/e-commerce/pkg/redis"
	"github.com/kareemhamed001/e-commerce/pkg/tracer"
	"github.com/kareemhamed001/e-commerce/services/CartService/config"
	"github.com/kareemhamed001/e-commerce/services/CartService/internal/delivery/grpc/handler"
	"github.com/kareemhamed001/e-commerce/services/CartService/internal/repository/redis"
	"github.com/kareemhamed001/e-commerce/services/CartService/internal/usecase"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracer := initTracing(ctx)
	defer shutdownTracer()

	redisCfg := &redisClient.Settings{
		RedisEnabled:  config.RedisEnabled,
		RedisHost:     config.RedisHost,
		RedisPort:     config.RedisPort,
		RedisPassword: config.RedisPassword,
		RedisDB:       config.RedisDB,
	}

	redisConn, err := redisClient.NewClientFromSettings(redisCfg)
	if err != nil {
		close(done)
		panic("failed to connect to redis")
	}

	productConn, err := grpc.Dial(config.ProductServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		close(done)
		panic("failed to connect to product service")
	}
	defer func() {
		_ = productConn.Close()
	}()

	userConn, err := grpc.Dial(config.UserServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		close(done)
		panic("failed to connect to user service")
	}
	defer func() {
		_ = userConn.Close()
	}()

	productClient := productpb.NewProductServiceClient(productConn)
	userClient := userpb.NewUserServiceClient(userConn)

	cartRepo := redis.NewCartRepository(redisConn)
	cartUsecase := usecase.NewCartUsecase(cartRepo, productClient, userClient, config.DownstreamTimeout)

	validate := validator.New()
	grpcHandler := handler.NewCartGRPCHandler(cartUsecase, validate)

	if err := grpcHandler.Run(done, config.GRPCPort); err != nil {
		logger.Errorf("failed to start gRPC server: %v", err)
		close(done)
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	close(done)
	_ = redisConn.Close()
	time.Sleep(200 * time.Millisecond)
}

func initTracing(ctx context.Context) func() {
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "ecommece_jaeger:4317")
	tp, err := tracer.InitTracer(ctx, "cart-service-grpc", jaegerEndpoint)
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
