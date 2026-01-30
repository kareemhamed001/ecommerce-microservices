package dto

type AddItemRequest struct {
	UserID    uint `json:"user_id" validate:"required,gt=0"`
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

type UpdateItemRequest struct {
	UserID    uint `json:"user_id" validate:"required,gt=0"`
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

type RemoveItemRequest struct {
	UserID    uint `json:"user_id" validate:"required,gt=0"`
	ProductID uint `json:"product_id" validate:"required,gt=0"`
}
