package clients

import (
	"fmt"

	"github.com/kareemhamed001/e-commerce/pkg/grpcmiddleware"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	cartpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/cart"
	orderpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/order"
	productpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/product"
	userpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceClients holds all gRPC client connections
type ServiceClients struct {
	UserClient    userpb.UserServiceClient
	ProductClient productpb.ProductServiceClient
	CartClient    cartpb.CartServiceClient
	OrderClient   orderpb.OrderServiceClient
	conns         []*grpc.ClientConn
}

// NewServiceClients creates new gRPC client connections to all services
func NewServiceClients(
	userServiceURL,
	productServiceURL,
	cartServiceURL,
	orderServiceURL,
	internalAuthToken string,
	cbConfig grpcmiddleware.CircuitBreakerConfig,
) (*ServiceClients, error) {
	clients := &ServiceClients{
		conns: make([]*grpc.ClientConn, 0),
	}

	// Connect to User Service
	userConn, err := createGRPCConnection(userServiceURL, internalAuthToken, cbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}
	clients.UserClient = userpb.NewUserServiceClient(userConn)
	clients.conns = append(clients.conns, userConn)
	logger.Infof("Connected to User Service at %s", userServiceURL)

	// Connect to Product Service
	productConn, err := createGRPCConnection(productServiceURL, internalAuthToken, cbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product service: %w", err)
	}
	clients.ProductClient = productpb.NewProductServiceClient(productConn)
	clients.conns = append(clients.conns, productConn)
	logger.Infof("Connected to Product Service at %s", productServiceURL)

	// Connect to Cart Service
	cartConn, err := createGRPCConnection(cartServiceURL, internalAuthToken, cbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cart service: %w", err)
	}
	clients.CartClient = cartpb.NewCartServiceClient(cartConn)
	clients.conns = append(clients.conns, cartConn)
	logger.Infof("Connected to Cart Service at %s", cartServiceURL)

	// Connect to Order Service
	orderConn, err := createGRPCConnection(orderServiceURL, internalAuthToken, cbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}
	clients.OrderClient = orderpb.NewOrderServiceClient(orderConn)
	clients.conns = append(clients.conns, orderConn)
	logger.Infof("Connected to Order Service at %s", orderServiceURL)

	return clients, nil
}

// createGRPCConnection creates a new gRPC connection with retry logic
func createGRPCConnection(target, internalAuthToken string, cbConfig grpcmiddleware.CircuitBreakerConfig) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcmiddleware.InternalAuthUnaryClientInterceptor(internalAuthToken),
			grpcmiddleware.CircuitBreakerUnaryClientInterceptor("api-gateway->"+target, cbConfig),
		),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
		),
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", target, err)
	}

	return conn, nil
}

// Close closes all gRPC connections
func (sc *ServiceClients) Close() error {
	for _, conn := range sc.conns {
		if err := conn.Close(); err != nil {
			logger.Errorf("Error closing gRPC connection: %v", err)
		}
	}
	logger.Info("All gRPC connections closed")
	return nil
}
