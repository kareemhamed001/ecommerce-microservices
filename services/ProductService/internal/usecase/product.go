package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	productCacheTTL     = 30 * time.Minute
	productListCacheTTL = 1 * time.Hour
)

type ProductUsecase struct {
	productRepo  domain.ProductRepository
	productCache domain.ProductCache
	tracer       trace.Tracer
}

var _ domain.ProductUsecase = (*ProductUsecase)(nil)

func NewProductUsecase(productRepo domain.ProductRepository, productCache domain.ProductCache) *ProductUsecase {
	return &ProductUsecase{
		productRepo:  productRepo,
		productCache: productCache,
		tracer:       otel.Tracer("product-usecase"),
	}
}

func (u *ProductUsecase) CreateProduct(ctx context.Context, productDto *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.CreateProduct")
	defer span.End()

	span.SetAttributes(
		attribute.String("product.name", productDto.Name),
		attribute.Float64("product.price", float64(productDto.Price)),
		attribute.Int("product.quantity", productDto.Quantity),
	)

	newProduct := &domain.Product{
		Name:             productDto.Name,
		ShortDescription: productDto.ShortDescription,
		Description:      productDto.Description,
		Price:            productDto.Price,
		DiscountType:     domain.DiscountType(productDto.DiscountType),
		DiscountValue:    productDto.DiscountValue,
		ImageUrl:         productDto.ImageUrl,
		Quantity:         productDto.Quantity,
	}

	_, dbSpan := u.tracer.Start(ctx, "Database.CreateProduct")
	if err := u.productRepo.CreateProduct(ctx, newProduct); err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	dbSpan.SetAttributes(attribute.Int("product.id", int(newProduct.ID)))
	dbSpan.End()

	span.SetStatus(codes.Ok, "Product created successfully")
	return &dto.ProductResponse{
		Id:               newProduct.ID,
		Name:             newProduct.Name,
		ShortDescription: newProduct.ShortDescription,
		Description:      newProduct.Description,
		Price:            newProduct.Price,
		DiscountType:     string(newProduct.DiscountType),
		DiscountValue:    newProduct.DiscountValue,
		ImageUrl:         newProduct.ImageUrl,
		Quantity:         newProduct.Quantity,
	}, nil
}

func (u *ProductUsecase) GetProductByID(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.GetProductByID")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", int(id)))

	_, cacheSpan := u.tracer.Start(ctx, "Cache.GetProduct")
	product, err := u.productCache.GetProduct(ctx, id)
	if err == nil {
		cacheSpan.SetAttributes(attribute.Bool("cache.hit", true))
		cacheSpan.End()
		logger.Debug("Product cache hit")
		span.SetAttributes(
			attribute.Bool("cache.hit", true),
			attribute.String("product.name", product.Name),
		)
		span.SetStatus(codes.Ok, "Product found in cache")
		return product, nil
	}
	cacheSpan.SetAttributes(attribute.Bool("cache.hit", false))
	cacheSpan.End()

	logger.Debug("Product cache miss, fetching from DB")
	_, dbSpan := u.tracer.Start(ctx, "Database.GetProductByID")
	productObj, err := u.productRepo.GetProductByID(ctx, id)
	if err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	dbSpan.End()

	newProduct := &dto.ProductResponse{
		Id:               productObj.ID,
		Name:             productObj.Name,
		ShortDescription: productObj.ShortDescription,
		Description:      productObj.Description,
		Price:            productObj.Price,
		DiscountType:     string(productObj.DiscountType),
		DiscountValue:    productObj.DiscountValue,
		ImageUrl:         productObj.ImageUrl,
		Quantity:         productObj.Quantity,
	}

	_, setCacheSpan := u.tracer.Start(ctx, "Cache.SetProduct")
	if err := u.productCache.SetProduct(ctx, newProduct, productCacheTTL); err != nil {
		setCacheSpan.RecordError(err)
		logger.Warnf("Failed to cache product: %v", err)
	}
	setCacheSpan.End()

	span.SetAttributes(
		attribute.Bool("cache.hit", false),
		attribute.String("product.name", newProduct.Name),
	)
	span.SetStatus(codes.Ok, "Product retrieved from database")
	return newProduct, nil
}

func (u *ProductUsecase) ListProducts(ctx context.Context, page, perPage int) ([]dto.ProductResponse, int, error) {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.ListProducts")
	defer span.End()

	_, dbSpan := u.tracer.Start(ctx, "Database.ListProducts")
	products, total, err := u.productRepo.ListProducts(ctx, page, perPage)
	if err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}
	dbSpan.SetAttributes(attribute.Int("products.count", len(products)))
	dbSpan.End()

	span.SetAttributes(attribute.Int("products.count", len(products)))
	span.SetStatus(codes.Ok, "Products retrieved from database")

	productsMapped := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		productsMapped[i] = dto.ProductResponse{
			Id:               p.ID,
			Name:             p.Name,
			ShortDescription: p.ShortDescription,
			Description:      p.Description,
			Price:            p.Price,
			DiscountType:     string(p.DiscountType),
			DiscountValue:    p.DiscountValue,
			ImageUrl:         p.ImageUrl,
			Quantity:         p.Quantity,
		}
	}

	return productsMapped, total, nil
}

func (u *ProductUsecase) UpdateProduct(ctx context.Context, id uint, product *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.UpdateProduct")
	defer span.End()

	span.SetAttributes(
		attribute.Int("product.id", int(id)),
		attribute.String("product.name", *product.Name),
		attribute.Float64("product.price", float64(*product.Price)),
	)

	newProduct := &domain.Product{
		Name:             *product.Name,
		ShortDescription: product.ShortDescription,
		Description:      *product.Description,
		Price:            *product.Price,
		DiscountType:     domain.DiscountType(*product.DiscountType),
		DiscountValue:    *product.DiscountValue,
		ImageUrl:         product.ImageUrl,
		Quantity:         *product.Quantity,
	}

	_, dbSpan := u.tracer.Start(ctx, "Database.UpdateProduct")
	if err := u.productRepo.UpdateProduct(ctx, id, newProduct); err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	dbSpan.End()

	_, deleteSpan := u.tracer.Start(ctx, "Cache.DeleteProduct")
	if err := u.productCache.DeleteProduct(ctx, id); err != nil {
		deleteSpan.RecordError(err)
		logger.Warnf("Failed to delete product from cache: %v", err)
	}
	deleteSpan.End()

	_, invalidateSpan := u.tracer.Start(ctx, "Cache.DeleteProduct")
	if err := u.productCache.DeleteProduct(ctx, id); err != nil {
		invalidateSpan.RecordError(err)
		logger.Warnf("Failed to delete product from cache: %v", err)
	}
	invalidateSpan.End()

	span.SetStatus(codes.Ok, "Product updated successfully")
	return nil, nil
}

func (u *ProductUsecase) RestockProduct(ctx context.Context, id uint, quantity int) error {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.RestockProduct")
	defer span.End()

	span.SetAttributes(
		attribute.Int("product.id", int(id)),
		attribute.Int("product.restock_quantity", quantity),
	)

	if quantity <= 0 {
		err := errors.New("quantity must be greater than zero")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	product, err := u.productRepo.GetProductByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	product.Quantity += quantity
	if err := u.productRepo.UpdateProduct(ctx, id, product); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "product restocked")
	return nil
}

func (u *ProductUsecase) DeleteProduct(ctx context.Context, id uint) error {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.DeleteProduct")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", int(id)))

	_, dbSpan := u.tracer.Start(ctx, "Database.DeleteProduct")
	if err := u.productRepo.DeleteProduct(ctx, id); err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	dbSpan.End()

	_, deleteSpan := u.tracer.Start(ctx, "Cache.DeleteProduct")
	if err := u.productCache.DeleteProduct(ctx, id); err != nil {
		deleteSpan.RecordError(err)
		logger.Warnf("Failed to delete product from cache: %v", err)
	}
	deleteSpan.End()

	_, invalidateSpan := u.tracer.Start(ctx, "Cache.DeleteProduct")
	if err := u.productCache.DeleteProduct(ctx, id); err != nil {
		invalidateSpan.RecordError(err)
		logger.Warnf("Failed to delete product from cache: %v", err)
	}
	invalidateSpan.End()

	span.SetStatus(codes.Ok, "Product deleted successfully")
	return nil
}
