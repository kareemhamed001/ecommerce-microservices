package domain

type Address struct {
	ID      uint   `gorm:"primaryKey;autoIncrement" json:"id" validate:"-"`
	UserID  uint   `gorm:"not null;index" json:"user_id" validate:"required"`
	Country string `gorm:"type:varchar(50);not null" json:"country" validate:"required,min=2,max=50"`
	City    string `gorm:"type:varchar(50);not null" json:"city" validate:"required,min=2,max=50"`
	State   string `gorm:"type:varchar(50);not null" json:"state" validate:"required,min=2,max=50"`
	Street  string `gorm:"type:varchar(100);not null" json:"street" validate:"required,min=2,max=100"`
	ZipCode string `gorm:"type:varchar(20);null" json:"zip_code" validate:"omitempty,min=2,max=20"`
}
