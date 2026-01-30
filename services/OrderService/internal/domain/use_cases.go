package domain

import (
	"context"

	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/delivery/grpc/dto"
)

type OrderUsecase interface {
	CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetOrderByID(ctx context.Context, id uint) (*dto.OrderResponse, error)
	ListOrders(ctx context.Context, userID *uint, page, perPage int) ([]dto.OrderResponse, int, error)
	AddOrderItem(ctx context.Context, req *dto.AddOrderItemRequest) (*dto.OrderResponse, error)
	RemoveOrderItem(ctx context.Context, orderID, itemID uint) (*dto.OrderResponse, error)
	UpdateOrderStatus(ctx context.Context, orderID uint, status string) (*dto.OrderResponse, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrderByID(ctx context.Context, id uint) (*Order, error)
	ListOrders(ctx context.Context, userID *uint, page, perPage int) ([]Order, int, error)
	AddOrderItem(ctx context.Context, item *OrderItem) error
	RemoveOrderItem(ctx context.Context, orderID, itemID uint) error
	UpdateOrderStatus(ctx context.Context, orderID uint, status OrderStatus) error
	UpdateOrderTotal(ctx context.Context, orderID uint, total float32) error
}