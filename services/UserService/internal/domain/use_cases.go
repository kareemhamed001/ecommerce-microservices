package domain

import (
	"context"

	"github.com/kareemhamed001/e-commerce/services/UserService/internal/delivery/grpc/dto"
)

type AddressUsecaseInterface interface {
	CreateAddress(ctx context.Context, req *dto.CreateAddressRequest) (int32, error)
	GetAddressByID(ctx context.Context, addressID int32) (Address, error)
	ListAddressesByUserID(ctx context.Context, userID int32) ([]Address, error)
	UpdateAddress(ctx context.Context, req *dto.UpdateAddressRequest) error
	DeleteAddress(ctx context.Context, addressID int32) error
}

type UserUsecaseInterface interface {
	Login(ctx context.Context, email, password string) (*dto.UserResponse, error)
	CreateUser(context.Context, *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(context.Context, uint) (*dto.UserResponse, error)
	GetUserByEmail(context.Context, string) (*dto.UserResponse, error)
	ListUsers(context.Context, int, int) ([]*dto.UserResponse, error)
	ListUsersByRole(context.Context, string, int, int) ([]*dto.UserResponse, error)
	SearchUsers(context.Context, string, int, int) ([]*dto.UserResponse, error)
	UpdateUser(context.Context, *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(context.Context, uint) error
}
