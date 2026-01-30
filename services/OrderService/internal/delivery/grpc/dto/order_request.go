package dto

type OrderItemInput struct {
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

type CreateOrderRequest struct {
	UserID               uint             `json:"user_id" validate:"required,gt=0"`
	ShippingCost         float32          `json:"shipping_cost" validate:"gte=0"`
	ShippingDurationDays int              `json:"shipping_duration_days" validate:"gte=0"`
	Discount             float32          `json:"discount" validate:"gte=0"`
	Items                []OrderItemInput `json:"items" validate:"required,min=1,dive"`
}

type AddOrderItemRequest struct {
	OrderID   uint `json:"order_id" validate:"required,gt=0"`
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

type UpdateOrderStatusRequest struct {
	OrderID uint   `json:"order_id" validate:"required,gt=0"`
	Status  string `json:"status" validate:"required,oneof=pending paid shipped delivered canceled"`
}