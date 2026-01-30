package dto

type CategoryResponse struct {
	Id          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}
