package dto

type CartItemResponse struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type CartResponse struct {
	UserID        uint               `json:"user_id"`
	Items         []CartItemResponse `json:"items"`
	TotalQuantity int                `json:"total_quantity"`
}
