package domain

import "gorm.io/gorm"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCanceled  OrderStatus = "canceled"
)

type Order struct {
	gorm.Model
	UserID               uint        `json:"user_id"`
	ShippingCost         float32     `json:"shipping_cost"`
	ShippingDurationDays int         `json:"shipping_duration_days"`
	Discount             float32     `json:"discount"`
	Total                float32     `json:"total"`
	Status               OrderStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Items                []OrderItem `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type OrderItem struct {
	gorm.Model
	OrderID    uint    `json:"order_id"`
	ProductID  uint    `json:"product_id"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float32 `json:"unit_price"`
	TotalPrice float32 `json:"total_price"`
}