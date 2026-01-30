package domain

import "time"

type Category struct {
	ID          uint    `gorm:"primarykey"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
