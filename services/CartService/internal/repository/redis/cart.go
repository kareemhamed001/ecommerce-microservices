package redis

import (
	"context"
	"fmt"
	"strconv"

	redisClient "github.com/kareemhamed001/e-commerce/pkg/redis"
	"github.com/kareemhamed001/e-commerce/services/CartService/internal/domain"
)

const cartKeyPrefix = "cart:"

type CartRepository struct {
	client *redisClient.Client
}

var _ domain.CartRepository = (*CartRepository)(nil)

func NewCartRepository(client *redisClient.Client) *CartRepository {
	return &CartRepository{client: client}
}

func (r *CartRepository) GetCart(ctx context.Context, userID uint) (domain.Cart, error) {
	if !r.client.IsEnabled() {
		return domain.Cart{}, fmt.Errorf("redis disabled")
	}

	key := cartKey(userID)
	values, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Cart{}, err
	}

	items := make([]domain.CartItem, 0, len(values))
	var totalQty int
	for productIDStr, qtyStr := range values {
		productID64, err := strconv.ParseUint(productIDStr, 10, 32)
		if err != nil {
			continue
		}
		qty, err := strconv.Atoi(qtyStr)
		if err != nil {
			continue
		}
		items = append(items, domain.CartItem{
			ProductID: uint(productID64),
			Quantity:  qty,
		})
		totalQty += qty
	}

	return domain.Cart{
		UserID:        userID,
		Items:         items,
		TotalQuantity: totalQty,
	}, nil
}

func (r *CartRepository) AddItem(ctx context.Context, userID, productID uint, quantity int) error {
	if !r.client.IsEnabled() {
		return fmt.Errorf("redis disabled")
	}

	key := cartKey(userID)
	return r.client.HIncrBy(ctx, key, fmt.Sprintf("%d", productID), int64(quantity)).Err()
}

func (r *CartRepository) UpdateItem(ctx context.Context, userID, productID uint, quantity int) error {
	if !r.client.IsEnabled() {
		return fmt.Errorf("redis disabled")
	}

	key := cartKey(userID)
	return r.client.HSet(ctx, key, fmt.Sprintf("%d", productID), quantity).Err()
}

func (r *CartRepository) RemoveItem(ctx context.Context, userID, productID uint) error {
	if !r.client.IsEnabled() {
		return fmt.Errorf("redis disabled")
	}

	key := cartKey(userID)
	return r.client.HDel(ctx, key, fmt.Sprintf("%d", productID)).Err()
}

func (r *CartRepository) ClearCart(ctx context.Context, userID uint) error {
	if !r.client.IsEnabled() {
		return fmt.Errorf("redis disabled")
	}

	key := cartKey(userID)
	return r.client.Del(ctx, key).Err()
}

func cartKey(userID uint) string {
	return fmt.Sprintf("%s%d", cartKeyPrefix, userID)
}
