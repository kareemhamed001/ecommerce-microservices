package domain

import (
	"context"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *Product) error
	GetProductByID(ctx context.Context, id uint) (*Product, error)
	GetProductsByIDs(ctx context.Context, ids []uint) ([]Product, error)
	UpdateProduct(ctx context.Context, id uint, product *Product) error
	ListProducts(ctx context.Context, page, perPage int) ([]Product, int, error)
	DeleteProduct(ctx context.Context, id uint) error
}

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *Category) error
	GetCategoryByID(ctx context.Context, id uint) (*Category, error)
	UpdateCategory(ctx context.Context, id uint, category *Category) error
	ListCategories(ctx context.Context, page, perPage int) ([]Category, int, error)
	DeleteCategory(ctx context.Context, id uint) error
}
