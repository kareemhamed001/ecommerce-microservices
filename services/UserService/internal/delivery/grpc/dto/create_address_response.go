package dto

type AddressResponse struct {
	ID      int32  `json:"id"`
	UserID  int32  `json:"user_id"`
	Country string `json:"country"`
	City    string `json:"city"`
	State   string `json:"state"`
	Street  string `json:"street"`
	ZipCode string `json:"zip_code"`
}
