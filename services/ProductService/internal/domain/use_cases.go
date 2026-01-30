package domain

import (
	"context"

	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/delivery/grpc/dto"
)

type ProductUsecase interface {
	CreateProduct(ctx context.Context, product *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProductByID(ctx context.Context, id uint) (*dto.ProductResponse, error)
	ListProducts(ctx context.Context, page, perPage int) ([]dto.ProductResponse, int, error)
	UpdateProduct(ctx context.Context, id uint, product *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id uint) error
	RestockProduct(ctx context.Context, id uint, quantity int) error
}

type CategoryUsecase interface {
	CreateCategory(ctx context.Context, category *dto.CreateCategoryRequest) error
	GetCategoryByID(ctx context.Context, id uint) (*dto.CategoryResponse, error)
	ListCategories(ctx context.Context, page, perPage int) ([]dto.CategoryResponse, int, error)
	UpdateCategory(ctx context.Context, id uint, category *dto.UpdateCategoryRequest) error
	DeleteCategory(ctx context.Context, id uint) error
}
