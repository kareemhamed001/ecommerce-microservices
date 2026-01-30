package domain

type UserRole string

const (
	AdminRole    UserRole = "admin"
	CustomerRole UserRole = "customer"
)

type User struct {
	ID       uint     `gorm:"primaryKey;autoIncrement" json:"id" validate:"-"`
	Name     string   `gorm:"type:varchar(100);not null" json:"name" validate:"required,min=2,max=100"`
	Email    string   `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" validate:"required,email"`
	Password string   `gorm:"type:varchar(255);not null" json:"password" validate:"required,min=6"`
	Role     UserRole `gorm:"type:varchar(50);not null" json:"role" validate:"required,oneof=admin customer"`
}
