package redisCache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisClient "github.com/kareemhamed001/e-commerce/pkg/redis"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/ProductService/internal/domain"
)

const (
	productKeyPrefix     = "product:"
	productListKeyPrefix = "products:list"
)

var _ domain.ProductCache = (*ProductCache)(nil)

type ProductCache struct {
	client *redisClient.Client
}

func NewProductCache(client *redisClient.Client) *ProductCache {
	return &ProductCache{client: client}
}

// GetProduct retrieves a product from cache by ID
func (c *ProductCache) GetProduct(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	if !c.client.IsEnabled() {
		return nil, fmt.Errorf("cache disabled")
	}

	key := fmt.Sprintf("%s%d", productKeyPrefix, id)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var product domain.Product
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}

	return &dto.ProductResponse{
		Id:               product.ID,
		Name:             product.Name,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		Price:            product.Price,
		Quantity:         product.Quantity,
		ImageUrl:         product.ImageUrl,
		DiscountType:     string(product.DiscountType),
		DiscountValue:    product.DiscountValue,
	}, nil
}

// SetProduct stores a product in cache
func (c *ProductCache) SetProduct(ctx context.Context, product *dto.ProductResponse, ttl time.Duration) error {
	if !c.client.IsEnabled() {
		return nil // Graceful degradation
	}

	key := fmt.Sprintf("%s%d", productKeyPrefix, product.Id)
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// DeleteProduct removes a product from cache
func (c *ProductCache) DeleteProduct(ctx context.Context, id uint) error {
	if !c.client.IsEnabled() {
		return nil
	}

	key := fmt.Sprintf("%s%d", productKeyPrefix, id)
	return c.client.Del(ctx, key).Err()
}

// GetProductList retrieves product list from cache
func (c *ProductCache) GetProductList(ctx context.Context, key string) ([]domain.Product, error) {
	if !c.client.IsEnabled() {
		return nil, fmt.Errorf("cache disabled")
	}

	cacheKey := fmt.Sprintf("%s:%s", productListKeyPrefix, key)
	data, err := c.client.Get(ctx, cacheKey).Bytes()
	if err != nil {
		return nil, err
	}

	var products []domain.Product
	if err := json.Unmarshal(data, &products); err != nil {
		return nil, err
	}

	return products, nil
}

// SetProductList stores product list in cache
func (c *ProductCache) SetProductList(ctx context.Context, key string, products []domain.Product, ttl time.Duration) error {
	if !c.client.IsEnabled() {
		return nil
	}

	cacheKey := fmt.Sprintf("%s:%s", productListKeyPrefix, key)
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, cacheKey, data, ttl).Err()
}

// InvalidateProductList clears all product list caches
func (c *ProductCache) InvalidateProductList(ctx context.Context) error {
	if !c.client.IsEnabled() {
		return nil
	}

	pattern := fmt.Sprintf("%s:*", productListKeyPrefix)
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
