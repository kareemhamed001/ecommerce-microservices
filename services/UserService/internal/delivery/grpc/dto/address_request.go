package dto

type CreateAddressRequest struct {
	UserID  int32  `json:"user_id" validate:"required"`
	Country string `json:"country" validate:"required"`
	City    string `json:"city" validate:"required"`
	State   string `json:"state" validate:"required"`
	Street  string `json:"street" validate:"required"`
	ZipCode string `json:"zip_code" validate:"required,len=5"`
}

type UpdateAddressRequest struct {
	Country string `json:"country" validate:"omitempty"`
	City    string `json:"city" validate:"omitempty"`
	State   string `json:"state" validate:"omitempty"`
	Street  string `json:"street" validate:"omitempty"`
	ZipCode string `json:"zip_code" validate:"omitempty,len=5"`
}
