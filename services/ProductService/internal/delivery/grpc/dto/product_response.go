package dto

type ProductResponse struct {
	Id               uint    `json:"id"`
	Name             string  `json:"name"`
	ShortDescription *string `json:"short_description,omitempty"`
	Description      string  `json:"description"`
	Price            float32 `json:"price"`
	DiscountType     string  `json:"discount_type"`
	DiscountValue    float32 `json:"discount_value"`
	ImageUrl         *string `json:"image_url,omitempty"`
	Quantity         int     `json:"quantity"`
}
