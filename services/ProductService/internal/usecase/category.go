package usecase

import (
	"context"

	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var _ domain.CategoryUsecase = (*CategoryUsecase)(nil)

type CategoryUsecase struct {
	categoryRepo domain.CategoryRepository
	tracer       trace.Tracer
}

func NewCategoryUsecase(categoryRepo domain.CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{
		categoryRepo: categoryRepo,
		tracer:       otel.Tracer("CategoryUsecase"),
	}
}

func (u *CategoryUsecase) CreateCategory(ctx context.Context, categoryDTO *dto.CreateCategoryRequest) error {
	ctx, span := u.tracer.Start(ctx, "CreateCategory")
	defer span.End()

	category := &domain.Category{
		Name:        categoryDTO.Name,
		Description: categoryDTO.Description,
	}

	err := u.categoryRepo.CreateCategory(ctx, category)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create category")
		return err
	}

	span.SetStatus(codes.Ok, "category created successfully")
	return nil
}
func (u *CategoryUsecase) GetCategoryByID(ctx context.Context, id uint) (*dto.CategoryResponse, error) {
	ctx, span := u.tracer.Start(ctx, "GetCategoryByID")
	defer span.End()

	category, err := u.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get category by ID")
		return nil, err
	}

	span.SetStatus(codes.Ok, "category retrieved successfully")
	return &dto.CategoryResponse{
		Id:          category.ID,
		Name:        category.Name,
		Description: category.Description,
	}, nil
}

func (u *CategoryUsecase) ListCategories(ctx context.Context, page, perPage int) ([]dto.CategoryResponse, int, error) {
	ctx, span := u.tracer.Start(ctx, "ListCategories")
	defer span.End()

	categories, total, err := u.categoryRepo.ListCategories(ctx, page, perPage)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list categories")
		return nil, 0, err
	}

	var categoryResponses []dto.CategoryResponse
	for _, category := range categories {
		categoryResponses = append(categoryResponses, dto.CategoryResponse{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
		})
	}

	span.SetStatus(codes.Ok, "categories listed successfully")
	return categoryResponses, total, nil
}

func (u *CategoryUsecase) UpdateCategory(ctx context.Context, id uint, categoryDTO *dto.UpdateCategoryRequest) error {
	ctx, span := u.tracer.Start(ctx, "UpdateCategory")
	defer span.End()

	category := &domain.Category{
		Name:        *categoryDTO.Name,
		Description: categoryDTO.Description,
	}

	err := u.categoryRepo.UpdateCategory(ctx, id, category)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update category")
		return err
	}

	span.SetStatus(codes.Ok, "category updated successfully")
	return nil
}

func (u *CategoryUsecase) DeleteCategory(ctx context.Context, id uint) error {
	ctx, span := u.tracer.Start(ctx, "DeleteCategory")
	defer span.End()

	err := u.categoryRepo.DeleteCategory(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete category")
		return err
	}

	span.SetStatus(codes.Ok, "category deleted successfully")
	return nil
}
