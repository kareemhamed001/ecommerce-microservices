package domain

import (
	"context"

	"github.com/kareemhamed001/e-commerce/services/CartService/internal/delivery/grpc/dto"
)

type CartUsecase interface {
	GetCart(ctx context.Context, userID uint) (*dto.CartResponse, error)
	AddItem(ctx context.Context, req *dto.AddItemRequest) (*dto.CartResponse, error)
	UpdateItem(ctx context.Context, req *dto.UpdateItemRequest) (*dto.CartResponse, error)
	RemoveItem(ctx context.Context, req *dto.RemoveItemRequest) (*dto.CartResponse, error)
	ClearCart(ctx context.Context, userID uint) error
}

type CartRepository interface {
	GetCart(ctx context.Context, userID uint) (Cart, error)
	AddItem(ctx context.Context, userID, productID uint, quantity int) error
	UpdateItem(ctx context.Context, userID, productID uint, quantity int) error
	RemoveItem(ctx context.Context, userID, productID uint) error
	ClearCart(ctx context.Context, userID uint) error
}
