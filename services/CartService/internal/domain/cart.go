package domain

type CartItem struct {
	ProductID uint
	Quantity  int
}

type Cart struct {
	UserID        uint
	Items         []CartItem
	TotalQuantity int
}
