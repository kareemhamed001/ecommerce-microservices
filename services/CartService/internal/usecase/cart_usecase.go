package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/kareemhamed001/e-commerce/services/CartService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/CartService/internal/domain"
	productpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/product"
	userpb "github.com/kareemhamed001/e-commerce/shared/proto/v1/user"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type CartUsecase struct {
	repo              domain.CartRepository
	productClient     productpb.ProductServiceClient
	userClient        userpb.UserServiceClient
	downstreamTimeout time.Duration
	tracer            trace.Tracer
}

var _ domain.CartUsecase = (*CartUsecase)(nil)

func NewCartUsecase(repo domain.CartRepository, productClient productpb.ProductServiceClient, userClient userpb.UserServiceClient, downstreamTimeout time.Duration) *CartUsecase {
	if downstreamTimeout <= 0 {
		downstreamTimeout = 3 * time.Second
	}

	return &CartUsecase{
		repo:              repo,
		productClient:     productClient,
		userClient:        userClient,
		downstreamTimeout: downstreamTimeout,
		tracer:            otel.Tracer("cart-usecase"),
	}
}

func (u *CartUsecase) GetCart(ctx context.Context, userID uint) (*dto.CartResponse, error) {
	ctx, span := u.tracer.Start(ctx, "CartUsecase.GetCart")
	defer span.End()

	if err := u.ensureUserExists(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	cart, err := u.repo.GetCart(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapCartToResponse(cart), nil
}

func (u *CartUsecase) AddItem(ctx context.Context, req *dto.AddItemRequest) (*dto.CartResponse, error) {
	ctx, span := u.tracer.Start(ctx, "CartUsecase.AddItem")
	defer span.End()

	span.SetAttributes(
		attribute.Int("cart.user_id", int(req.UserID)),
		attribute.Int("cart.product_id", int(req.ProductID)),
	)

	if err := u.ensureUserExists(ctx, req.UserID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if _, err := u.ensureProductExists(ctx, req.ProductID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := u.repo.AddItem(ctx, req.UserID, req.ProductID, req.Quantity); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	cart, err := u.repo.GetCart(ctx, req.UserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapCartToResponse(cart), nil
}

func (u *CartUsecase) UpdateItem(ctx context.Context, req *dto.UpdateItemRequest) (*dto.CartResponse, error) {
	ctx, span := u.tracer.Start(ctx, "CartUsecase.UpdateItem")
	defer span.End()

	if err := u.ensureUserExists(ctx, req.UserID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if _, err := u.ensureProductExists(ctx, req.ProductID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := u.repo.UpdateItem(ctx, req.UserID, req.ProductID, req.Quantity); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	cart, err := u.repo.GetCart(ctx, req.UserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapCartToResponse(cart), nil
}

func (u *CartUsecase) RemoveItem(ctx context.Context, req *dto.RemoveItemRequest) (*dto.CartResponse, error) {
	ctx, span := u.tracer.Start(ctx, "CartUsecase.RemoveItem")
	defer span.End()

	if err := u.ensureUserExists(ctx, req.UserID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := u.repo.RemoveItem(ctx, req.UserID, req.ProductID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	cart, err := u.repo.GetCart(ctx, req.UserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapCartToResponse(cart), nil
}

func (u *CartUsecase) ClearCart(ctx context.Context, userID uint) error {
	ctx, span := u.tracer.Start(ctx, "CartUsecase.ClearCart")
	defer span.End()

	if err := u.ensureUserExists(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if err := u.repo.ClearCart(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (u *CartUsecase) ensureUserExists(ctx context.Context, userID uint) error {
	ctx, cancel := context.WithTimeout(ctx, u.downstreamTimeout)
	defer cancel()

	_, err := u.userClient.GetUserByID(ctx, &userpb.GetUserByIDRequest{Id: int32(userID)})
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	return nil
}

func (u *CartUsecase) ensureProductExists(ctx context.Context, productID uint) (*productpb.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, u.downstreamTimeout)
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

func mapCartToResponse(cart domain.Cart) *dto.CartResponse {
	items := make([]dto.CartItemResponse, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, dto.CartItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return &dto.CartResponse{
		UserID:        cart.UserID,
		Items:         items,
		TotalQuantity: cart.TotalQuantity,
	}
}
