package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/e-commerce/pkg/grpcmiddleware"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/services/ApiGateway/config"
	"github.com/kareemhamed001/e-commerce/services/ApiGateway/internal/clients"
	"github.com/kareemhamed001/e-commerce/services/ApiGateway/internal/handlers"
	"github.com/kareemhamed001/e-commerce/services/ApiGateway/internal/router"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.InitGlobal("development", "logs/gateway/system.log")
		logger.Errorf("Failed to load configuration: %v", err)
		return
	}

	// Initialize logger
	logger.InitGlobal(cfg.AppEnv, "logs/gateway/system.log")
	logger.Info("event=startup component=api-gateway message=starting")
	logger.Info("event=config_loaded component=api-gateway message=configuration loaded")

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize gRPC clients
	serviceClients, err := clients.NewServiceClients(
		cfg.UserServiceURL,
		cfg.ProductServiceURL,
		cfg.CartServiceURL,
		cfg.OrderServiceURL,
		cfg.InternalAuthToken,
		grpcmiddleware.CircuitBreakerConfig{
			Enabled:      cfg.CircuitBreakerEnabled,
			MaxRequests:  cfg.CircuitBreakerMaxRequests,
			Interval:     cfg.CircuitBreakerInterval,
			Timeout:      cfg.CircuitBreakerTimeout,
			FailureRatio: cfg.CircuitBreakerFailureRatio,
			MinRequests:  cfg.CircuitBreakerMinRequests,
		},
	)
	if err != nil {
		logger.Errorf("Failed to initialize service clients: %v", err)
		return
	}
	var closeOnce sync.Once
	closeClients := func() {
		closeOnce.Do(func() {
			logger.Info("event=shutdown_step component=grpc_clients action=close")
			serviceClients.Close()
		})
	}
	defer closeClients()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(serviceClients.UserClient)
	productHandler := handlers.NewProductHandler(serviceClients.ProductClient)
	cartHandler := handlers.NewCartHandler(serviceClients.CartClient)
	orderHandler := handlers.NewOrderHandler(serviceClients.OrderClient)

	routerEngine := gin.Default()

	// Initialize router
	apiRouter := router.NewRouter(routerEngine, cfg, userHandler, productHandler, cartHandler, orderHandler)

	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer baseCancel()

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      apiRouter.Handler(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		// Ensure handlers can derive a base context that is canceled on shutdown.
		BaseContext: func(_ net.Listener) context.Context {
			return baseCtx
		},
	}
	server.RegisterOnShutdown(func() {
		closeClients()
	})

	serverErr := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		logger.Infof("event=server_start component=http_server addr=:%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				serverErr <- nil
				return
			}
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	// Wait for interrupt signal or server error for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		logger.Infof("event=shutdown_start component=api-gateway reason=signal signal=%s", sig.String())
	case err := <-serverErr:
		if err != nil {
			logger.Errorf("event=server_error component=http_server error=%v", err)
		}
		logger.Info("event=server_stopped component=http_server")
		return
	}

	// Graceful shutdown with timeout
	shutdownTimeout := 30 * time.Second
	logger.Infof("event=shutdown_timeout component=http_server timeout=%s", shutdownTimeout)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Stop accepting new connections immediately
	logger.Info("event=shutdown_step component=http_server action=disable_keepalives")
	server.SetKeepAlivesEnabled(false)
	logger.Info("event=shutdown_step component=http_server action=cancel_base_context")
	baseCancel()
	logger.Info("event=shutdown_step component=http_server action=shutdown")

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("event=shutdown_error component=http_server error=%v", err)
	}

	closeClients()

	// Ensure the server goroutine has completed
	if err := <-serverErr; err != nil {
		logger.Errorf("event=shutdown_error component=http_server error=%v", err)
	}

	logger.Info("event=shutdown_complete component=api-gateway")
}
