package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/domain"
	productpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/product"
	userpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/user"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	downstreamTimeout = 3 * time.Second
)

type OrderUsecase struct {
	orderRepo     domain.OrderRepository
	productClient productpb.ProductServiceClient
	userClient    userpb.UserServiceClient
	tracer        trace.Tracer
}

var _ domain.OrderUsecase = (*OrderUsecase)(nil)

func NewOrderUsecase(orderRepo domain.OrderRepository, productClient productpb.ProductServiceClient, userClient userpb.UserServiceClient) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:     orderRepo,
		productClient: productClient,
		userClient:    userClient,
		tracer:        otel.Tracer("order-usecase"),
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.CreateOrder")
	defer span.End()

	span.SetAttributes(attribute.Int("order.user_id", int(req.UserID)))

	if err := u.ensureUserExists(ctx, req.UserID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	items := make([]domain.OrderItem, 0, len(req.Items))
	var itemsTotal float32

	for _, item := range req.Items {
		product, err := u.ensureProductExists(ctx, item.ProductID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		unitPrice := product.GetPrice()
		totalPrice := unitPrice * float32(item.Quantity)
		itemsTotal += totalPrice

		items = append(items, domain.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  unitPrice,
			TotalPrice: totalPrice,
		})
	}

	total := calculateOrderTotal(itemsTotal, req.ShippingCost, req.Discount)

	order := &domain.Order{
		UserID:               req.UserID,
		ShippingCost:         req.ShippingCost,
		ShippingDurationDays: req.ShippingDurationDays,
		Discount:             req.Discount,
		Total:                total,
		Status:               domain.OrderStatusPending,
		Items:                items,
	}

	if err := u.orderRepo.CreateOrder(ctx, order); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("order.id", int(order.ID)))
	span.SetStatus(codes.Ok, "order created")
	return mapOrderToResponse(order), nil
}

func (u *OrderUsecase) GetOrderByID(ctx context.Context, id uint) (*dto.OrderResponse, error) {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.GetOrderByID")
	defer span.End()

	order, err := u.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "order fetched")
	return mapOrderToResponse(order), nil
}

func (u *OrderUsecase) ListOrders(ctx context.Context, userID *uint, page, perPage int) ([]dto.OrderResponse, int, error) {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.ListOrders")
	defer span.End()

	orders, total, err := u.orderRepo.ListOrders(ctx, userID, page, perPage)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}

	response := make([]dto.OrderResponse, 0, len(orders))
	for i := range orders {
		response = append(response, *mapOrderToResponse(&orders[i]))
	}

	span.SetStatus(codes.Ok, "orders listed")
	return response, total, nil
}

func (u *OrderUsecase) AddOrderItem(ctx context.Context, req *dto.AddOrderItemRequest) (*dto.OrderResponse, error) {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.AddOrderItem")
	defer span.End()

	product, err := u.ensureProductExists(ctx, req.ProductID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	unitPrice := product.GetPrice()
	item := &domain.OrderItem{
		OrderID:    req.OrderID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		UnitPrice:  unitPrice,
		TotalPrice: unitPrice * float32(req.Quantity),
	}

	if err := u.orderRepo.AddOrderItem(ctx, item); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	order, err := u.orderRepo.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	itemsTotal := sumItemsTotal(order.Items)
	updatedTotal := calculateOrderTotal(itemsTotal, order.ShippingCost, order.Discount)
	if err := u.orderRepo.UpdateOrderTotal(ctx, order.ID, updatedTotal); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	order.Total = updatedTotal

	return mapOrderToResponse(order), nil
}

func (u *OrderUsecase) RemoveOrderItem(ctx context.Context, orderID, itemID uint) (*dto.OrderResponse, error) {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.RemoveOrderItem")
	defer span.End()

	if err := u.orderRepo.RemoveOrderItem(ctx, orderID, itemID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	order, err := u.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	itemsTotal := sumItemsTotal(order.Items)
	updatedTotal := calculateOrderTotal(itemsTotal, order.ShippingCost, order.Discount)
	if err := u.orderRepo.UpdateOrderTotal(ctx, order.ID, updatedTotal); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	order.Total = updatedTotal

	return mapOrderToResponse(order), nil
}

func (u *OrderUsecase) UpdateOrderStatus(ctx context.Context, orderID uint, status string) (*dto.OrderResponse, error) {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.UpdateOrderStatus")
	defer span.End()

	orderStatus := domain.OrderStatus(status)
	if err := u.orderRepo.UpdateOrderStatus(ctx, orderID, orderStatus); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	order, err := u.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapOrderToResponse(order), nil
}

func (u *OrderUsecase) ensureUserExists(ctx context.Context, userID uint) error {
	ctx, cancel := context.WithTimeout(ctx, downstreamTimeout)
	defer cancel()

	_, err := u.userClient.GetUserByID(ctx, &userpb.GetUserByIDRequest{Id: int32(userID)})
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	return nil
}

func (u *OrderUsecase) ensureProductExists(ctx context.Context, productID uint) (*productpb.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, downstreamTimeout)
	defer cancel()

	response, err := u.productClient.GetProductByID(ctx, &productpb.GetProductByIDRequest{Id: int64(productID)})
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}
	if response.GetProduct() == nil {
		return nil, fmt.Errorf("product not found: empty response")
	}
	return response.GetProduct(), nil
}

func mapOrderToResponse(order *domain.Order) *dto.OrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, dto.OrderItemResponse{
			ID:         item.ID,
			OrderID:    item.OrderID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		})
	}

	return &dto.OrderResponse{
		ID:               order.ID,
		UserID:           order.UserID,
		ShippingCost:     order.ShippingCost,
		ShippingDuration: order.ShippingDurationDays,
		Discount:         order.Discount,
		Total:            order.Total,
		Status:           string(order.Status),
		Items:            items,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}
}

func sumItemsTotal(items []domain.OrderItem) float32 {
	var total float32
	for _, item := range items {
		if item.TotalPrice > 0 {
			total += item.TotalPrice
			continue
		}
		total += item.UnitPrice * float32(item.Quantity)
	}
	return total
}

func calculateOrderTotal(itemsTotal, shippingCost, discount float32) float32 {
	if discount < 0 {
		discount = 0
	}
	if shippingCost < 0 {
		shippingCost = 0
	}
	total := itemsTotal + shippingCost - discount
	if total < 0 {
		return 0
	}
	return total
}
