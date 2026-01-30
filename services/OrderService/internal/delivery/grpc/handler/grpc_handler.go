package handler

import (
	"context"
	"net"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/domain"
	orderpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/order"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type OrderGRPCHandler struct {
	orderpb.UnimplementedOrderServiceServer
	orderUsecase domain.OrderUsecase
	validate     *validator.Validate
	tracer       trace.Tracer
}

var _ orderpb.OrderServiceServer = (*OrderGRPCHandler)(nil)

func NewOrderGRPCHandler(orderUsecase domain.OrderUsecase, validate *validator.Validate) *OrderGRPCHandler {
	return &OrderGRPCHandler{
		orderUsecase: orderUsecase,
		validate:     validate,
		tracer:       otel.Tracer("order_GRPC_handler"),
	}
}

func (h *OrderGRPCHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	reqCtx, span := h.tracer.Start(ctx, "OrderHandler.CreateOrder")
	defer span.End()

	items := make([]dto.OrderItemInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, dto.OrderItemInput{
			ProductID: uint(item.GetProductId()),
			Quantity:  int(item.GetQuantity()),
		})
	}

	createReq := dto.CreateOrderRequest{
		UserID:               uint(req.GetUserId()),
		ShippingCost:         req.GetShippingCost(),
		ShippingDurationDays: int(req.GetShippingDurationDays()),
		Discount:             req.GetDiscount(),
		Items:                items,
	}

	if err := h.validate.Struct(&createReq); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return nil, err
	}

	order, err := h.orderUsecase.CreateOrder(reqCtx, &createReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("order.id", int(order.ID)))
	return &orderpb.CreateOrderResponse{Order: mapOrderToPB(order)}, nil
}

func (h *OrderGRPCHandler) GetOrderByID(ctx context.Context, req *orderpb.GetOrderByIDRequest) (*orderpb.GetOrderByIDResponse, error) {
	reqCtx, span := h.tracer.Start(ctx, "OrderHandler.GetOrderByID")
	defer span.End()

	order, err := h.orderUsecase.GetOrderByID(reqCtx, uint(req.GetId()))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &orderpb.GetOrderByIDResponse{Order: mapOrderToPB(order)}, nil
}

func (h *OrderGRPCHandler) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	reqCtx, span := h.tracer.Start(ctx, "OrderHandler.ListOrders")
	defer span.End()

	page := int(req.GetPage())
	if page == 0 {
		page = 1
	}
	perPage := int(req.GetPerPage())
	if perPage == 0 {
		perPage = 10
	}

	var userID *uint
	if req.GetUserId() > 0 {
		id := uint(req.GetUserId())
		userID = &id
	}

	orders, total, err := h.orderUsecase.ListOrders(reqCtx, userID, page, perPage)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	responseOrders := make([]*orderpb.Order, 0, len(orders))
	for i := range orders {
		responseOrders = append(responseOrders, mapOrderToPB(&orders[i]))
	}

	return &orderpb.ListOrdersResponse{
		Orders:     responseOrders,
		TotalCount: int32(total),
	}, nil
}

func (h *OrderGRPCHandler) AddOrderItem(ctx context.Context, req *orderpb.AddOrderItemRequest) (*orderpb.AddOrderItemResponse, error) {
	reqCtx, span := h.tracer.Start(ctx, "OrderHandler.AddOrderItem")
	defer span.End()

	addReq := dto.AddOrderItemRequest{
		OrderID:   uint(req.GetOrderId()),
		ProductID: uint(req.GetProductId()),
		Quantity:  int(req.GetQuantity()),
	}

	if err := h.validate.Struct(&addReq); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return nil, err
	}

	order, err := h.orderUsecase.AddOrderItem(reqCtx, &addReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &orderpb.AddOrderItemResponse{Order: mapOrderToPB(order)}, nil
}

func (h *OrderGRPCHandler) RemoveOrderItem(ctx context.Context, req *orderpb.RemoveOrderItemRequest) (*orderpb.RemoveOrderItemResponse, error) {
	reqCtx, span := h.tracer.Start(ctx, "OrderHandler.RemoveOrderItem")
	defer span.End()

	order, err := h.orderUsecase.RemoveOrderItem(reqCtx, uint(req.GetOrderId()), uint(req.GetItemId()))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &orderpb.RemoveOrderItemResponse{Order: mapOrderToPB(order)}, nil
}

func (h *OrderGRPCHandler) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	reqCtx, span := h.tracer.Start(ctx, "OrderHandler.UpdateOrderStatus")
	defer span.End()

	updateReq := dto.UpdateOrderStatusRequest{
		OrderID: uint(req.GetOrderId()),
		Status:  req.GetStatus(),
	}

	if err := h.validate.Struct(&updateReq); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return nil, err
	}

	order, err := h.orderUsecase.UpdateOrderStatus(reqCtx, updateReq.OrderID, updateReq.Status)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &orderpb.UpdateOrderStatusResponse{Order: mapOrderToPB(order)}, nil
}

func (h *OrderGRPCHandler) Run(done <-chan any, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Errorf("Error while starting order grpc server: %v", err)
		return err
	}

	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, h)

	go func() {
		logger.Infof("Order gRPC server is running on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Errorf("Error while serving order grpc server: %v", err)
		}
	}()

	go func() {
		<-done
		logger.Info("Shutting down order gRPC server...")
		grpcServer.GracefulStop()
	}()

	return nil
}

func mapOrderToPB(order *dto.OrderResponse) *orderpb.Order {
	if order == nil {
		return nil
	}

	items := make([]*orderpb.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &orderpb.OrderItem{
			Id:         int64(item.ID),
			OrderId:    int64(item.OrderID),
			ProductId:  int64(item.ProductID),
			Quantity:   int32(item.Quantity),
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		})
	}

	return &orderpb.Order{
		Id:                   int64(order.ID),
		UserId:               int64(order.UserID),
		ShippingCost:         order.ShippingCost,
		ShippingDurationDays: int32(order.ShippingDuration),
		Discount:             order.Discount,
		Total:                order.Total,
		Status:               order.Status,
		Items:                items,
		CreatedAt:            formatTime(order.CreatedAt),
		UpdatedAt:            formatTime(order.UpdatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
