package handler

import (
	"context"
	"net"

	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/e-commerce/pkg/jwt"
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/domain"
	pb "github.com/kareemhamed001/e-commerce/shared/proto/v1/user"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type UserGRPCHandler struct {
	pb.UnimplementedUserServiceServer
	userUsecase    domain.UserUsecaseInterface
	addressUsecase domain.AddressUsecaseInterface
	validate       *validator.Validate
	jwtManager     *jwt.JWTManager
	tracer         trace.Tracer
}

func NewUserGRPCHandler(userUsecase domain.UserUsecaseInterface, addressUsecase domain.AddressUsecaseInterface, validate *validator.Validate, jwtManager *jwt.JWTManager) *UserGRPCHandler {
	return &UserGRPCHandler{
		userUsecase:    userUsecase,
		addressUsecase: addressUsecase,
		validate:       validate,
		jwtManager:     jwtManager,
		tracer:         otel.Tracer("user_GRPC_handler"),
	}
}

// CreateUser(ctx context.Context, in *CreateUserRequest, opts ...grpc.CallOption) (*CreateUserResponse, error)
func (h *UserGRPCHandler) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.CreateUser")
	defer span.End()

	_, validationSpan := h.tracer.Start(ctx, "Validate CreateUserRequest")

	createUserRequestDto := dto.CreateUserRequest{
		Name:     in.GetName(),
		Email:    in.GetEmail(),
		Password: in.GetPassword(),
	}

	err := h.validate.Struct(createUserRequestDto)
	if err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, err.Error())
		validationSpan.End()
		return nil, err
	}
	validationSpan.End()

	createUserCtx, createUserSpan := h.tracer.Start(ctx, "Usecase CreateUser")

	createUserResponse, err := h.userUsecase.CreateUser(createUserCtx, &createUserRequestDto)
	if err != nil {
		createUserSpan.RecordError(err)
		createUserSpan.SetStatus(codes.Error, err.Error())
		createUserSpan.End()
		return nil, err
	}
	createUserSpan.End()
	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:    int32(createUserResponse.ID),
			Name:  createUserResponse.Name,
			Email: createUserResponse.Email,
			Role:  createUserResponse.Role,
		},
	}, nil
}

func (h *UserGRPCHandler) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.Login")
	defer span.End()

	_, validationSpan := h.tracer.Start(ctx, "Validate LoginRequest")

	loginRequestDto := dto.LoginRequest{
		Email:    in.GetEmail(),
		Password: in.GetPassword(),
	}
	err := h.validate.Struct(loginRequestDto)
	if err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, err.Error())
		validationSpan.End()
		return nil, err
	}
	validationSpan.End()

	loginCtx, loginSpan := h.tracer.Start(ctx, "Usecase Login")

	userResponse, err := h.userUsecase.Login(loginCtx, loginRequestDto.Email, loginRequestDto.Password)
	if err != nil {
		err = domain.ErrInvalidCredentials
		loginSpan.RecordError(err)
		loginSpan.SetStatus(codes.Error, err.Error())
		loginSpan.End()
		return nil, err
	}
	loginSpan.End()

	_, jwtSpan := h.tracer.Start(ctx, "Generate JWT Token")
	token, err := h.jwtManager.Generate(userResponse.ID, userResponse.Email, userResponse.Role)
	if err != nil {
		jwtSpan.RecordError(err)
		jwtSpan.SetStatus(codes.Error, err.Error())
		jwtSpan.End()
		return nil, err
	}
	jwtSpan.End()

	return &pb.LoginResponse{
		Token: token,
	}, nil
}

func (h *UserGRPCHandler) GetUserByID(ctx context.Context, in *pb.GetUserByIDRequest) (*pb.User, error) {
	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.GetUserByID")
	defer span.End()

	userId := in.GetId()

	userResponse, err := h.userUsecase.GetUserByID(ctx, uint(userId))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &pb.User{
		Id:    int32(userResponse.ID),
		Name:  userResponse.Name,
		Email: userResponse.Email,
		Role:  userResponse.Role,
	}, nil
}

func (h *UserGRPCHandler) SearchUsers(ctx context.Context, in *pb.SearchUsersRequest) (*pb.SearchUsersResponse, error) {
	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.SearchUsers")
	defer span.End()

	query := in.GetQuery()
	page := in.GetPageNumber()
	limit := in.GetPageSize()

	_, searchUsersSpan := h.tracer.Start(ctx, "Usecase SearchUsers")

	usersResponse, err := h.userUsecase.SearchUsers(ctx, query, int(page), int(limit))
	if err != nil {
		searchUsersSpan.RecordError(err)
		searchUsersSpan.SetStatus(codes.Error, err.Error())
		searchUsersSpan.End()
		return nil, err
	}
	searchUsersSpan.End()

	_, mapSpan := h.tracer.Start(ctx, "Map UsersResponse to pb.Users")
	pbUsers := make([]*pb.User, len(usersResponse))
	for i, user := range usersResponse {

		pbUsers[i] = &pb.User{
			Id:    int32(user.ID),
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		}
	}
	mapSpan.End()

	return &pb.SearchUsersResponse{
		Users: pbUsers,
	}, nil
}

func (h *UserGRPCHandler) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.User, error) {
	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.UpdateUser")
	defer span.End()

	updateUserRequest := dto.UpdateUserRequest{
		Name:     in.GetName(),
		Email:    in.GetEmail(),
		Password: in.GetPassword(),
	}

	_, validationSpan := h.tracer.Start(ctx, "Validate UpdateUserRequest")

	err := h.validate.Struct(updateUserRequest)
	if err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, err.Error())
		validationSpan.End()
		return nil, err
	}
	validationSpan.End()

	updateUserCtx, updateUserSpan := h.tracer.Start(ctx, "Usecase UpdateUser")

	userResponse, err := h.userUsecase.UpdateUser(updateUserCtx, &updateUserRequest)
	if err != nil {
		updateUserSpan.RecordError(err)
		updateUserSpan.SetStatus(codes.Error, err.Error())
		updateUserSpan.End()
		return nil, err
	}
	updateUserSpan.End()

	return &pb.User{
		Id:    int32(userResponse.ID),
		Name:  userResponse.Name,
		Email: userResponse.Email,
		Role:  userResponse.Role,
	}, nil
}

func (h *UserGRPCHandler) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.DeleteUser")
	defer span.End()

	userId := in.GetId()

	err := h.userUsecase.DeleteUser(ctx, uint(userId))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return &pb.DeleteUserResponse{Success: false}, err
	}
	return &pb.DeleteUserResponse{Success: true}, nil
}

func (h *UserGRPCHandler) CreateAddress(ctx context.Context, in *pb.CreateAddressRequest) (*pb.CreateAddressResponse, error) {
	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.CreateAddress")
	defer span.End()

	addressRequest := dto.CreateAddressRequest{
		UserID:  in.GetUserId(),
		Country: in.GetCountry(),
		City:    in.GetCity(),
		State:   in.GetState(),
		Street:  in.GetStreet(),
		ZipCode: in.GetZipCode(),
	}

	_, validationSpan := h.tracer.Start(ctx, "Validate CreateAddressRequest")

	err := h.validate.Struct(addressRequest)
	if err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, err.Error())
		validationSpan.End()
		return nil, err
	}
	validationSpan.End()

	createAddressCtx, createAddressSpan := h.tracer.Start(ctx, "Usecase CreateAddress")

	_, err = h.addressUsecase.CreateAddress(createAddressCtx, &addressRequest)
	if err != nil {
		createAddressSpan.RecordError(err)
		createAddressSpan.SetStatus(codes.Error, err.Error())
		createAddressSpan.End()
		return nil, err
	}
	createAddressSpan.End()

	return &pb.CreateAddressResponse{}, nil
}
func (h *UserGRPCHandler) GetAddressByID(ctx context.Context, in *pb.GetAddressByIDRequest) (*pb.GetAddressByIDResponse, error) {
	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.GetAddressByID")
	defer span.End()

	addressId := in.GetId()

	getAddressCtx, getAddressSpan := h.tracer.Start(ctx, "Usecase GetAddressByID")

	_, err := h.addressUsecase.GetAddressByID(getAddressCtx, addressId)
	if err != nil {
		getAddressSpan.RecordError(err)
		getAddressSpan.SetStatus(codes.Error, err.Error())
		getAddressSpan.End()
		return nil, err
	}
	getAddressSpan.End()

	return &pb.GetAddressByIDResponse{}, nil
}
func (h *UserGRPCHandler) ListAddressesByUserID(ctx context.Context, in *pb.ListAddressesByUserIDRequest) (*pb.ListAddressesByUserIDResponse, error) {

	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.ListAddressesByUserID")
	defer span.End()

	userId := in.GetUserId()

	listAddressesCtx, listAddressesSpan := h.tracer.Start(ctx, "Usecase ListAddressesByUserID")

	_, err := h.addressUsecase.ListAddressesByUserID(listAddressesCtx, userId)
	if err != nil {
		listAddressesSpan.RecordError(err)
		listAddressesSpan.SetStatus(codes.Error, err.Error())
		listAddressesSpan.End()
		return nil, err
	}
	listAddressesSpan.End()

	return &pb.ListAddressesByUserIDResponse{}, nil
}
func (h *UserGRPCHandler) UpdateAddress(ctx context.Context, in *pb.UpdateAddressRequest) (*pb.UpdateAddressResponse, error) {

	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.UpdateAddress")
	defer span.End()

	_, validateAddressSpan := h.tracer.Start(ctx, "Validate UpdateAddressRequest")

	updateAddressRequest := dto.UpdateAddressRequest{
		Country: in.GetCountry(),
		City:    in.GetCity(),
		State:   in.GetState(),
		Street:  in.GetStreet(),
		ZipCode: in.GetZipCode(),
	}

	err := h.validate.Struct(updateAddressRequest)
	if err != nil {
		validateAddressSpan.RecordError(err)
		validateAddressSpan.SetStatus(codes.Error, err.Error())
		validateAddressSpan.End()
		return nil, err
	}
	validateAddressSpan.End()

	updateAddressCtx, updateAddressSpan := h.tracer.Start(ctx, "Usecase UpdateAddress")

	err = h.addressUsecase.UpdateAddress(updateAddressCtx, &updateAddressRequest)
	if err != nil {
		updateAddressSpan.RecordError(err)
		updateAddressSpan.SetStatus(codes.Error, err.Error())
		updateAddressSpan.End()
		return nil, err
	}
	updateAddressSpan.End()

	return &pb.UpdateAddressResponse{}, nil
}
func (h *UserGRPCHandler) DeleteAddress(ctx context.Context, in *pb.DeleteAddressRequest) (*pb.DeleteAddressResponse, error) {
	ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.DeleteAddress")
	defer span.End()

	addressId := in.GetId()

	deleteAddressCtx, deleteAddressSpan := h.tracer.Start(ctx, "Usecase DeleteAddress")

	err := h.addressUsecase.DeleteAddress(deleteAddressCtx, addressId)
	if err != nil {
		deleteAddressSpan.RecordError(err)
		deleteAddressSpan.SetStatus(codes.Error, err.Error())
		deleteAddressSpan.End()
		return nil, err
	}
	deleteAddressSpan.End()

	return &pb.DeleteAddressResponse{}, nil
}

func (h *UserGRPCHandler) Run(done <-chan any, port string) error {
	// Implementation here
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Errorf("Error while starting user grpc server: %v", err)
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, h)

	go func() {
		logger.Infof("User gRPC server is running on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Errorf("Error while serving user grpc server: %v", err)
		}
	}()

	go func() {
		<-done
		logger.Info("Shutting down user gRPC server...")
		grpcServer.GracefulStop()
	}()

	return nil
}
