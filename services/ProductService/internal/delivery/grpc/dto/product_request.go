package dto

type CreateProductRequest struct {
	Name              string  `json:"name" validate:"required,min=2,max=100"`
	ShortDescription  *string `json:"short_description" validate:"omitempty,min=2,max=150"`
	Description       string  `json:"description" validate:"required,min=2"`
	Price             float32 `json:"price" validate:"required,gt=0"`
	DiscountType      string  `json:"discount_type" validate:"omitempty,oneof=fixed percent"`
	DiscountValue     float32 `json:"discount_value" validate:"omitempty,gt=0"`
	DiscountStartDate *string `json:"discount_start_date" validate:"omitempty,datetime=2006-01-02"`
	DiscountEndDate   *string `json:"discount_end_date" validate:"omitempty,datetime=2006-01-02"`
	ImageUrl          *string `json:"image_url" validate:"omitempty,url"`
	Quantity          int     `json:"quantity" validate:"required,gte=0"`
}

type UpdateProductRequest struct {
	Name              *string  `json:"name" validate:"omitempty,min=2,max=100"`
	ShortDescription  *string  `json:"short_description" validate:"omitempty,min=2,max=150"`
	Description       *string  `json:"description" validate:"omitempty,min=2"`
	Price             *float32 `json:"price" validate:"omitempty,gt=0"`
	DiscountType      *string  `json:"discount_type" validate:"omitempty,oneof=fixed percent"`
	DiscountValue     *float32 `json:"discount_value" validate:"omitempty,gt=0"`
	DiscountStartDate *string  `json:"discount_start_date" validate:"omitempty,datetime=2006-01-02"`
	DiscountEndDate   *string  `json:"discount_end_date" validate:"omitempty,datetime=2006-01-02"`
	ImageUrl          *string  `json:"image_url" validate:"omitempty,url"`
	Quantity          *int     `json:"quantity" validate:"omitempty,gte=0"`
}
