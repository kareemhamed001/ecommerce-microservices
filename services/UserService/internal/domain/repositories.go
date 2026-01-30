package domain

import (
	"context"
)

type UserRepositoryInterface interface {
	CreateUser(context.Context, *User) (User, error)
	GetUserByID(context.Context, uint) (User, error)
	GetUserByEmail(context.Context, string) (User, error)
	ListUsers(context.Context, int, int) ([]User, error)
	ListUsersByRole(context.Context, UserRole, int, int) ([]User, error)
	SearchUsers(context.Context, string, int, int) ([]User, error)
	UpdateUser(context.Context, User) (User, error)
	DeleteUser(context.Context, uint) error
}

type AddressRepositoryInterface interface {
	CreateAddress(context.Context, *Address) (Address, error)
	GetAddressByID(context.Context, uint) (Address, error)
	ListAddressesByUserID(context.Context, uint, int, int) ([]Address, error)
	UpdateAddress(context.Context, Address) (Address, error)
	DeleteAddress(context.Context, uint) error
}
