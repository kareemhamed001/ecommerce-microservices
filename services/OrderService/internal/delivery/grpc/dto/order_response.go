package dto

import "time"

type OrderItemResponse struct {
	ID         uint    `json:"id"`
	OrderID    uint    `json:"order_id"`
	ProductID  uint    `json:"product_id"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float32 `json:"unit_price"`
	TotalPrice float32 `json:"total_price"`
}

type OrderResponse struct {
	ID               uint                `json:"id"`
	UserID           uint                `json:"user_id"`
	ShippingCost     float32             `json:"shipping_cost"`
	ShippingDuration int                 `json:"shipping_duration_days"`
	Discount         float32             `json:"discount"`
	Total            float32             `json:"total"`
	Status           string              `json:"status"`
	Items            []OrderItemResponse `json:"items"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}
