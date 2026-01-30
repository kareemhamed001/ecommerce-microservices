package usecase

import (
	"context"

	"github.com/kareemhamed001/e-commerce/pkg/password"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// type UserUsecaseInterface interface {
// 	Login(ctx context.Context, email, password string) (*dto.UserResponse, error)
// 	CreateUser(context.Context, *dto.CreateUserRequest) (*dto.UserResponse, error)
// 	GetUserByID(context.Context, uint) (*dto.UserResponse, error)
// 	GetUserByEmail(context.Context, string) (*dto.UserResponse, error)
// 	ListUsers(context.Context, int, int) ([]*dto.UserResponse, error)
// 	ListUsersByRole(context.Context, string, int, int) ([]*dto.UserResponse, error)
// 	SearchUsers(context.Context, string, int, int) ([]*dto.UserResponse, error)
// 	UpdateUser(context.Context, *dto.UpdateUserRequest) (*dto.UserResponse, error)
// 	DeleteUser(context.Context, uint) error
// }

type UserUsecase struct {
	userRepo domain.UserRepositoryInterface
	tracer   trace.Tracer
}

func NewUserUsecase(userRepo domain.UserRepositoryInterface) domain.UserUsecaseInterface {
	return &UserUsecase{
		userRepo: userRepo,
		tracer:   otel.Tracer("user_usecase"),
	}
}

func (u *UserUsecase) Login(ctx context.Context, email, passwords string) (*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.Login")
	defer span.End()

	gettinUserByEmailCtx, gettingUserByEmailSpan := u.tracer.Start(ctx, "userRepo.GetUserByEmail")
	user, err := u.userRepo.GetUserByEmail(gettinUserByEmailCtx, email)
	if err != nil {
		gettingUserByEmailSpan.RecordError(err)
		gettingUserByEmailSpan.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	gettingUserByEmailSpan.End()

	_, validatePasswordSpan := u.tracer.Start(ctx, "password.Verify")

	valid := password.Verify(passwords, user.Password)
	if !valid {
		err := domain.ErrInvalidCredentials
		validatePasswordSpan.RecordError(err)
		validatePasswordSpan.SetStatus(codes.Error, err.Error())
		validatePasswordSpan.End()

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	validatePasswordSpan.End()

	return nil, nil
}

func (u *UserUsecase) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.CreateUser")
	defer span.End()

	_, hashingPasswordSpan := u.tracer.Start(ctx, "password.Hash")

	hashedPassword, err := password.Hash(req.Password)
	if err != nil {

		hashingPasswordSpan.RecordError(err)
		hashingPasswordSpan.SetStatus(codes.Error, err.Error())
		hashingPasswordSpan.End()

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, domain.ErrHashingPassword
	}
	hashingPasswordSpan.End()

	createUserCtx, createUserSpan := u.tracer.Start(ctx, "userRepo.CreateUser")

	user, err := u.userRepo.CreateUser(createUserCtx, &domain.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Role:     domain.CustomerRole,
	})
	if err != nil {

		createUserSpan.RecordError(err)
		createUserSpan.SetStatus(codes.Error, err.Error())
		createUserSpan.End()

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	createUserSpan.End()
	return &dto.UserResponse{
		ID:    uint(user.ID),
		Email: user.Email,
		Name:  user.Name,
		Role:  string(user.Role),
	}, nil
}

func (u *UserUsecase) GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.GetUserByID")
	defer span.End()

	span.SetAttributes(attribute.Int64("user_id", int64(id)))

	user, err := u.userRepo.GetUserByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  string(user.Role),
	}, nil
}

func (u *UserUsecase) GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.GetUserByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("email", email))

	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  string(user.Role),
	}, nil
}

func (u *UserUsecase) ListUsers(ctx context.Context, limit, offset int) ([]*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.ListUsers")
	defer span.End()

	span.SetAttributes(attribute.Int("limit", limit), attribute.Int("offset", offset))

	users, err := u.userRepo.ListUsers(ctx, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  string(user.Role),
		}
	}

	return userResponses, nil
}

func (u *UserUsecase) ListUsersByRole(ctx context.Context, role string, limit, offset int) ([]*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.ListUsersByRole")
	defer span.End()

	span.SetAttributes(attribute.String("role", role), attribute.Int("limit", limit), attribute.Int("offset", offset))

	userRole := domain.UserRole(role)
	users, err := u.userRepo.ListUsersByRole(ctx, userRole, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  string(user.Role),
		}
	}

	return userResponses, nil
}

func (u *UserUsecase) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.SearchUsers")
	defer span.End()

	span.SetAttributes(attribute.String("query", query), attribute.Int("limit", limit), attribute.Int("offset", offset))

	users, err := u.userRepo.SearchUsers(ctx, query, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  string(user.Role),
		}
	}

	return userResponses, nil
}

func (u *UserUsecase) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.UpdateUser")
	defer span.End()

	userToUpdate := domain.User{
		Name:  req.Name,
		Email: req.Email,
	}

	if req.Password != "" {
		_, hashingPasswordSpan := u.tracer.Start(ctx, "password.Hash")

		hashedPassword, err := password.Hash(req.Password)
		if err != nil {
			hashingPasswordSpan.RecordError(err)
			hashingPasswordSpan.SetStatus(codes.Error, err.Error())
			hashingPasswordSpan.End()

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, domain.ErrHashingPassword
		}
		hashingPasswordSpan.End()

		userToUpdate.Password = hashedPassword
	}

	updateUserCtx, updateUserSpan := u.tracer.Start(ctx, "userRepo.UpdateUser")

	user, err := u.userRepo.UpdateUser(updateUserCtx, userToUpdate)
	if err != nil {
		updateUserSpan.RecordError(err)
		updateUserSpan.SetStatus(codes.Error, err.Error())
		updateUserSpan.End()

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	updateUserSpan.End()

	return &dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  string(user.Role),
	}, nil
}

func (u *UserUsecase) DeleteUser(ctx context.Context, id uint) error {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.DeleteUser")
	defer span.End()

	span.SetAttributes(attribute.Int64("user_id", int64(id)))

	err := u.userRepo.DeleteUser(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}
