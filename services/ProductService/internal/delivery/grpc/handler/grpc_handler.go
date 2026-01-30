package handler

import (
	"context"
	"net"

	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/domain"
	pb "github.com/kareemhamed001/e-commerce/shared/proto/v1/product"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type ProductGRPCHandler struct {
	pb.UnimplementedProductServiceServer
	productUsecase  domain.ProductUsecase
	categoryUsecase domain.CategoryUsecase
	validate        *validator.Validate
	tracer          trace.Tracer
}

var _ pb.ProductServiceServer = (*ProductGRPCHandler)(nil)

func NewProductGRPCHandler(productUsecase domain.ProductUsecase, categoryUsecase domain.CategoryUsecase, validate *validator.Validate) *ProductGRPCHandler {
	return &ProductGRPCHandler{
		productUsecase:  productUsecase,
		categoryUsecase: categoryUsecase,
		validate:        validate,
		tracer:          otel.Tracer("product_GRPC_handler"),
	}
}

func (h *ProductGRPCHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	reqCtx, span := h.tracer.Start(ctx, "ProductHandler.CreateProduct")
	defer span.End()

	shortDesc := req.GetShortDescription()
	imageUrl := req.GetImageUrl()

	var discountType string
	switch req.GetDiscountType() {
	case pb.DiscountType_DISCOUNT_PERCENT:
		discountType = "percent"
	case pb.DiscountType_DISCOUNT_FIXED:
		discountType = "fixed"
	default:
		discountType = ""
	}

	productRequestDto := dto.CreateProductRequest{
		Name:             req.GetName(),
		ShortDescription: &shortDesc,
		Description:      req.GetDescription(),
		Price:            req.GetPrice(),
		DiscountType:     discountType,
		DiscountValue:    req.GetDiscountValue(),
		ImageUrl:         &imageUrl,
		Quantity:         int(req.GetQuantity()),
	}

	_, validationSpan := h.tracer.Start(reqCtx, "ProductHandler.ValidateProduct")
	if err := h.validate.Struct(&productRequestDto); err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, "validation failed")
		validationSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")

		return nil, err
	}
	validationSpan.End()

	span.SetAttributes(
		attribute.String("product.name", productRequestDto.Name),
		attribute.Float64("product.price", float64(productRequestDto.Price)),
		attribute.String("product.discount_type", productRequestDto.DiscountType),
	)
	product, err := h.productUsecase.CreateProduct(reqCtx, &productRequestDto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("product.id", int(product.Id)))
	productResponse := &pb.Product{
		Id:               int32(product.Id),
		Name:             product.Name,
		ShortDescription: *product.ShortDescription,
		Description:      product.Description,
		Price:            product.Price,
		DiscountType:     product.DiscountType,
		DiscountValue:    product.DiscountValue,
		ImageUrl:         *product.ImageUrl,
		Quantity:         int32(product.Quantity),
	}

	span.SetStatus(codes.Ok, "Product created successfully")
	return &pb.CreateProductResponse{
		Product: productResponse,
	}, nil
}

func (h *ProductGRPCHandler) GetProductByID(ctx context.Context, req *pb.GetProductByIDRequest) (*pb.GetProductByIDResponse, error) {
	id := req.GetId()
	reqCtx, span := h.tracer.Start(ctx, "ProductHandler.GetProduct")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", int(id)))

	product, err := h.productUsecase.GetProductByID(reqCtx, uint(id))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.String("product.name", product.Name),
		attribute.Float64("product.price", float64(product.Price)),
	)

	productResponse := &pb.Product{
		Id:               int32(product.Id),
		Name:             product.Name,
		ShortDescription: *product.ShortDescription,
		Description:      product.Description,
		Price:            product.Price,
		DiscountType:     string(product.DiscountType),
		DiscountValue:    product.DiscountValue,
		ImageUrl:         *product.ImageUrl,
		Quantity:         int32(product.Quantity),
	}

	span.SetAttributes(attribute.String("product.response", productResponse.String()))

	span.SetStatus(codes.Ok, "Product retrieved successfully")

	return &pb.GetProductByIDResponse{
		Product: productResponse,
	}, nil
}

func (h *ProductGRPCHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	// Implementation here

	reqCtx, span := h.tracer.Start(ctx, "ProductHandler.ListProducts")
	defer span.End()

	page := int(req.GetPage())
	if page == 0 {
		page = 1
	}
	limit := int(req.GetPerPage())
	if limit == 0 {
		limit = 10
	}

	span.SetAttributes(
		attribute.Int("pagination.page", page),
		attribute.Int("pagination.limit", limit),
	)

	products, total, err := h.productUsecase.ListProducts(reqCtx, page, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	span.SetAttributes(attribute.Int("products.count", len(products)))
	span.SetAttributes(attribute.Int("products.total", total))

	productResponse := make([]*pb.Product, 0, len(products))

	for _, p := range products {
		productResponse = append(productResponse, &pb.Product{
			Id:               int32(p.Id),
			Name:             p.Name,
			ShortDescription: *p.ShortDescription,
			Description:      p.Description,
			Price:            p.Price,
			DiscountType:     string(p.DiscountType),
			DiscountValue:    p.DiscountValue,
			ImageUrl:         *p.ImageUrl,
			Quantity:         int32(p.Quantity),
		})
	}

	span.SetStatus(codes.Ok, "Products retrieved successfully")

	return &pb.ListProductsResponse{
		Products:   productResponse,
		TotalCount: int32(total),
	}, nil
}

func (h *ProductGRPCHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	id := int(req.GetId())
	reqCtx, span := h.tracer.Start(ctx, "ProductHandler.UpdateProduct")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", id))

	name := req.GetName()
	shortDesc := req.GetShortDescription()
	description := req.GetDescription()
	price := req.GetPrice()
	discountValue := req.GetDiscountValue()
	imageUrl := req.GetImageUrl()
	quantity := int(req.GetQuantity())

	var discountType string
	switch req.GetDiscountType() {
	case pb.DiscountType_DISCOUNT_PERCENT:
		discountType = "percent"
	case pb.DiscountType_DISCOUNT_FIXED:
		discountType = "fixed"
	default:
		discountType = ""
	}

	productRequest := dto.UpdateProductRequest{
		Name:             &name,
		ShortDescription: &shortDesc,
		Description:      &description,
		Price:            &price,
		DiscountType:     &discountType,
		DiscountValue:    &discountValue,
		ImageUrl:         &imageUrl,
		Quantity:         &quantity,
	}

	_, validationSpan := h.tracer.Start(reqCtx, "ProductHandler.ValidateUpdateProduct")
	if err := h.validate.Struct(&productRequest); err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, "validation failed")
		validationSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")

		return nil, err
	}
	validationSpan.End()

	_, err := h.productUsecase.GetProductByID(reqCtx, uint(id))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "product not found")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("product.name", *productRequest.Name),
		attribute.Float64("product.price", float64(*productRequest.Price)),
	)
	productResponse, err := h.productUsecase.UpdateProduct(reqCtx, uint(id), &productRequest)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err

	}

	span.SetStatus(codes.Ok, "Product updated successfully")
	return &pb.UpdateProductResponse{
		Product: &pb.Product{
			Id:               int32(productResponse.Id),
			Name:             productResponse.Name,
			ShortDescription: *productResponse.ShortDescription,
			Description:      productResponse.Description,
			Price:            productResponse.Price,
			DiscountType:     string(productResponse.DiscountType),
			DiscountValue:    productResponse.DiscountValue,
			ImageUrl:         *productResponse.ImageUrl,
			Quantity:         int32(productResponse.Quantity),
		},
	}, nil
}

func (h *ProductGRPCHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	id := req.GetId()
	reqCtx, span := h.tracer.Start(ctx, "ProductHandler.DeleteProduct")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", int(id)))
	if err := h.productUsecase.DeleteProduct(reqCtx, uint(id)); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "Product deleted successfully")

	return &pb.DeleteProductResponse{
		Success: true,
	}, nil
}

// CreateCategory(context.Context, *CreateCategoryRequest) (*CreateCategoryResponse, error)
func (h *ProductGRPCHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	ctx, span := h.tracer.Start(ctx, "ProductHandler.CreateCategory")
	defer span.End()

	description := req.GetDescription()

	categoryDto := dto.CreateCategoryRequest{
		Name:        req.GetName(),
		Description: &description,
	}

	// Validation and creation logic here
	_, validationSpan := h.tracer.Start(ctx, "ProductHandler.ValidateCategory")
	if err := h.validate.Struct(&categoryDto); err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, "validation failed")
		validationSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")

		return nil, err
	}
	validationSpan.End()

	span.SetAttributes(
		attribute.String("category.name", categoryDto.Name),
	)
	err := h.categoryUsecase.CreateCategory(ctx, &categoryDto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "Category created successfully")

	return &pb.CreateCategoryResponse{
		Success: true,
		Message: "Category created successfully",
	}, nil
}

// GetCategoryByID(context.Context, *GetCategoryByIDRequest) (*GetCategoryByIDResponse, error)
func (h *ProductGRPCHandler) GetCategoryByID(ctx context.Context, req *pb.GetCategoryByIDRequest) (*pb.GetCategoryByIDResponse, error) {
	ctx, span := h.tracer.Start(ctx, "ProductHandler.GetCategoryByID")
	defer span.End()

	id := req.GetId()

	span.SetAttributes(attribute.Int("category.id", int(id)))

	category, err := h.categoryUsecase.GetCategoryByID(ctx, uint(id))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.String("category.name", category.Name),
	)

	span.SetStatus(codes.Ok, "Category retrieved successfully")

	return &pb.GetCategoryByIDResponse{
		Category: &pb.Category{
			Name:        category.Name,
			Description: *category.Description,
		},
	}, nil
}

// ListCategories(context.Context, *ListCategoriesRequest) (*ListCategoriesResponse, error)
func (h *ProductGRPCHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	ctx, span := h.tracer.Start(ctx, "ProductHandler.ListCategories")
	defer span.End()

	page := int(req.GetPage())
	perPage := int(req.GetPerPage())

	span.SetAttributes(
		attribute.Int("pagination.page", page),
		attribute.Int("pagination.per_page", perPage),
	)

	categories, total, err := h.categoryUsecase.ListCategories(ctx, page, perPage)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetAttributes(attribute.Int("categories.count", len(categories)))
	span.SetAttributes(attribute.Int("categories.total", total))

	var categoryResponses []*pb.Category
	for _, c := range categories {
		categoryResponses = append(categoryResponses, &pb.Category{
			Name:        c.Name,
			Description: *c.Description,
		})
	}

	span.SetStatus(codes.Ok, "Categories listed successfully")

	return &pb.ListCategoriesResponse{
		Categories: categoryResponses,
		TotalCount: int32(total),
	}, nil
}

// UpdateCategory(context.Context, *UpdateCategoryRequest) (*UpdateCategoryResponse, error)
func (h *ProductGRPCHandler) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	ctx, span := h.tracer.Start(ctx, "ProductHandler.UpdateCategory")
	defer span.End()

	updateDto := dto.UpdateCategoryRequest{
		Name:        &req.Name,
		Description: &req.Description,
	}

	// Validation and update logic here
	_, validationSpan := h.tracer.Start(ctx, "ProductHandler.ValidateUpdateCategory")
	if err := h.validate.Struct(&updateDto); err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, "validation failed")
		validationSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")

		return &pb.UpdateCategoryResponse{
			Success: false,
			Message: "Validation failed",
		}, err
	}
	validationSpan.End()

	span.SetAttributes(
		attribute.Int("category.id", int(req.GetId())),
		attribute.String("category.name", *updateDto.Name),
	)

	err := h.categoryUsecase.UpdateCategory(ctx, uint(req.GetId()), &updateDto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "Category updated successfully")

	return &pb.UpdateCategoryResponse{
		Success: true,
		Message: "Category updated successfully",
	}, nil
}

// DeleteCategory(context.Context, *DeleteCategoryRequest) (*DeleteCategoryResponse, error)
func (h *ProductGRPCHandler) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	ctx, span := h.tracer.Start(ctx, "ProductHandler.DeleteCategory")
	defer span.End()

	id := req.GetId()

	span.SetAttributes(attribute.Int("category.id", int(id)))

	err := h.categoryUsecase.DeleteCategory(ctx, uint(id))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "Category deleted successfully")
	return &pb.DeleteCategoryResponse{
		Success: true,
	}, nil
}

func (h *ProductGRPCHandler) Run(done <-chan any, port string) error {
	// Implementation here
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Errorf("Error while starting product grpc server: %v", err)
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, h)

	go func() {
		logger.Infof("Product gRPC server is running on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Errorf("Error while serving product grpc server: %v", err)
		}
	}()

	go func() {
		<-done
		logger.Info("Shutting down product gRPC server...")
		grpcServer.GracefulStop()
	}()

	return nil
}
