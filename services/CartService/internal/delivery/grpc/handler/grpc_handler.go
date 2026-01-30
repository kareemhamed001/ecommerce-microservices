package handler

import (
	"context"
	"net"

	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/services/CartService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/CartService/internal/domain"
	cartpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/cart"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type CartGRPCHandler struct {
	cartpb.UnimplementedCartServiceServer
	usecase  domain.CartUsecase
	validate *validator.Validate
	tracer   trace.Tracer
}

var _ cartpb.CartServiceServer = (*CartGRPCHandler)(nil)

func NewCartGRPCHandler(usecase domain.CartUsecase, validate *validator.Validate) *CartGRPCHandler {
	return &CartGRPCHandler{
		usecase:  usecase,
		validate: validate,
		tracer:   otel.Tracer("cart_GRPC_handler"),
	}
}

func (h *CartGRPCHandler) GetCart(ctx context.Context, req *cartpb.GetCartRequest) (*cartpb.CartResponse, error) {
	ctx, span := h.tracer.Start(ctx, "CartHandler.GetCart")
	defer span.End()

	userID := uint(req.GetUserId())
	response, err := h.usecase.GetCart(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapCartResponse(response), nil
}

func (h *CartGRPCHandler) AddItem(ctx context.Context, req *cartpb.AddItemRequest) (*cartpb.CartResponse, error) {
	ctx, span := h.tracer.Start(ctx, "CartHandler.AddItem")
	defer span.End()

	addReq := dto.AddItemRequest{
		UserID:    uint(req.GetUserId()),
		ProductID: uint(req.GetProductId()),
		Quantity:  int(req.GetQuantity()),
	}

	if err := h.validate.Struct(&addReq); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return nil, err
	}

	response, err := h.usecase.AddItem(ctx, &addReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapCartResponse(response), nil
}

func (h *CartGRPCHandler) UpdateItem(ctx context.Context, req *cartpb.UpdateItemRequest) (*cartpb.CartResponse, error) {
	ctx, span := h.tracer.Start(ctx, "CartHandler.UpdateItem")
	defer span.End()

	updateReq := dto.UpdateItemRequest{
		UserID:    uint(req.GetUserId()),
		ProductID: uint(req.GetProductId()),
		Quantity:  int(req.GetQuantity()),
	}

	if err := h.validate.Struct(&updateReq); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return nil, err
	}

	response, err := h.usecase.UpdateItem(ctx, &updateReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapCartResponse(response), nil
}

func (h *CartGRPCHandler) RemoveItem(ctx context.Context, req *cartpb.RemoveItemRequest) (*cartpb.CartResponse, error) {
	ctx, span := h.tracer.Start(ctx, "CartHandler.RemoveItem")
	defer span.End()

	removeReq := dto.RemoveItemRequest{
		UserID:    uint(req.GetUserId()),
		ProductID: uint(req.GetProductId()),
	}

	if err := h.validate.Struct(&removeReq); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return nil, err
	}

	response, err := h.usecase.RemoveItem(ctx, &removeReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapCartResponse(response), nil
}

func (h *CartGRPCHandler) ClearCart(ctx context.Context, req *cartpb.ClearCartRequest) (*cartpb.ClearCartResponse, error) {
	ctx, span := h.tracer.Start(ctx, "CartHandler.ClearCart")
	defer span.End()

	if err := h.usecase.ClearCart(ctx, uint(req.GetUserId())); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &cartpb.ClearCartResponse{Success: true}, nil
}

func (h *CartGRPCHandler) Run(done <-chan any, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Errorf("Error while starting cart grpc server: %v", err)
		return err
	}

	grpcServer := grpc.NewServer()
	cartpb.RegisterCartServiceServer(grpcServer, h)

	go func() {
		logger.Infof("Cart gRPC server is running on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Errorf("Error while serving cart grpc server: %v", err)
		}
	}()

	go func() {
		<-done
		logger.Info("Shutting down cart gRPC server...")
		grpcServer.GracefulStop()
	}()

	return nil
}

func mapCartResponse(response *dto.CartResponse) *cartpb.CartResponse {
	if response == nil {
		return &cartpb.CartResponse{}
	}

	items := make([]*cartpb.CartItem, 0, len(response.Items))
	for _, item := range response.Items {
		items = append(items, &cartpb.CartItem{
			ProductId: int64(item.ProductID),
			Quantity:  int32(item.Quantity),
		})
	}

	return &cartpb.CartResponse{
		UserId:        int64(response.UserID),
		Items:         items,
		TotalQuantity: int32(response.TotalQuantity),
	}
}
