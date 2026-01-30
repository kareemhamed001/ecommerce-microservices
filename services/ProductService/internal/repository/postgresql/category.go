package postgresql

import (
	"context"

	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

type CategoryRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		db:     db,
		tracer: otel.Tracer("CategoryRepository"),
	}
}

func (r *CategoryRepository) CreateCategory(ctx context.Context, category *domain.Category) error {
	ctx, span := r.tracer.Start(ctx, "CreateCategory")
	defer span.End()

	err := gorm.G[domain.Category](r.db).Create(ctx, category)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create category")
		return err
	}

	span.SetStatus(codes.Ok, "category created successfully")
	return nil

}
func (r *CategoryRepository) GetCategoryByID(ctx context.Context, id uint) (*domain.Category, error) {
	ctx, span := r.tracer.Start(ctx, "GetCategoryByID")
	defer span.End()

	category, err := gorm.G[domain.Category](r.db).
		Where("id = ?", id).
		First(ctx)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get category by ID")
		return nil, err
	}

	span.SetStatus(codes.Ok, "category retrieved successfully")
	return &category, nil

}
func (r *CategoryRepository) UpdateCategory(ctx context.Context, id uint, category *domain.Category) error {
	ctx, span := r.tracer.Start(ctx, "UpdateCategory")
	defer span.End()

	rowsAffected, err := gorm.G[domain.Category](r.db).
		Where("id = ?", id).
		Updates(ctx, *category)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update category")
		return err
	}
	if rowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		span.RecordError(err)
		span.SetStatus(codes.Error, "category not found")
		return err
	}

	span.SetStatus(codes.Ok, "category updated successfully")
	return nil

}
func (r *CategoryRepository) ListCategories(ctx context.Context, page, perPage int) ([]domain.Category, int, error) {

	ctx, span := r.tracer.Start(ctx, "ListCategories")
	defer span.End()

	categories, err := gorm.G[domain.Category](r.db).
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list categories")
		return nil, 0, err
	}

	total, err := gorm.G[domain.Category](r.db).
		Count(ctx, "*")

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to count categories")
		return nil, 0, err
	}

	span.SetStatus(codes.Ok, "categories listed successfully")
	return categories, int(total), nil
}
func (r *CategoryRepository) DeleteCategory(ctx context.Context, id uint) error {
	ctx, span := r.tracer.Start(ctx, "DeleteCategory")
	defer span.End()

	rowsAffected, err := gorm.G[domain.Category](r.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete category")
		return err
	}
	if rowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		span.RecordError(err)
		span.SetStatus(codes.Error, "category not found")
		return err
	}

	span.SetStatus(codes.Ok, "category deleted successfully")
	return nil

}
