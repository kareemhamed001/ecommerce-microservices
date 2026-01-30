package domain

import (
	"context"
	"time"

	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/delivery/grpc/dto"
)

type ProductCache interface {
	GetProduct(ctx context.Context, id uint) (*dto.ProductResponse, error)
	SetProduct(ctx context.Context, product *dto.ProductResponse, ttl time.Duration) error
	DeleteProduct(ctx context.Context, id uint) error
}
