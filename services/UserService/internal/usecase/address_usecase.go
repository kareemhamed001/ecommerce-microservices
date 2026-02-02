package usecase

import (
	"context"

	"github.com/kareemhamed001/e-commerce/services/UserService/internal/delivery/grpc/dto"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AddressUsecase struct {
	addressRepo domain.AddressRepositoryInterface
	userRepo    domain.UserRepositoryInterface
	tracer      trace.Tracer
}

var _ domain.AddressUsecaseInterface = (*AddressUsecase)(nil)

func NewAddressUsecase(addressRepo domain.AddressRepositoryInterface, userRepo domain.UserRepositoryInterface) domain.AddressUsecaseInterface {
	return &AddressUsecase{

		addressRepo: addressRepo,
		userRepo:    userRepo,
		tracer:      otel.Tracer("address_usecase"),
	}
}

func (a *AddressUsecase) CreateAddress(ctx context.Context, req *dto.CreateAddressRequest) (int32, error) {
	ctx, span := a.tracer.Start(ctx, "AddressUsecase.CreateAddress")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user_id", int64(req.UserID)),
		attribute.String("country", req.Country),
		attribute.String("city", req.City),
	)

	userExistsCtx, userExistsSpan := a.tracer.Start(ctx, "userRepo.UserExists")
	existedUser, err := a.userRepo.GetUserByID(userExistsCtx, uint(req.UserID))
	if err != nil {
		userExistsSpan.RecordError(err)
		userExistsSpan.SetStatus(codes.Error, err.Error())
		userExistsSpan.End()

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	userExistsSpan.End()

	if existedUser.ID == 0 {
		err := domain.ErrUserNotFound
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	createAddressCtx, createAddressSpan := a.tracer.Start(ctx, "addressRepo.CreateAddress")

	address, err := a.addressRepo.CreateAddress(createAddressCtx, &domain.Address{
		UserID:  uint(req.UserID),
		Country: req.Country,
		City:    req.City,
		State:   req.State,
		Street:  req.Street,
		ZipCode: req.ZipCode,
	})
	if err != nil {
		createAddressSpan.RecordError(err)
		createAddressSpan.SetStatus(codes.Error, err.Error())
		createAddressSpan.End()

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	createAddressSpan.End()

	return int32(address.ID), nil
}

func (a *AddressUsecase) GetAddressByID(ctx context.Context, addressID int32) (*dto.AddressResponse, error) {
	ctx, span := a.tracer.Start(ctx, "AddressUsecase.GetAddressByID")
	defer span.End()

	span.SetAttributes(
		attribute.Int("address_id", int(addressID)),
	)

	address, err := a.addressRepo.GetAddressByID(ctx, uint(addressID))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	response := dto.AddressResponse{
		ID:      int32(address.ID),
		UserID:  int32(address.UserID),
		Country: address.Country,
		City:    address.City,
		State:   address.State,
		Street:  address.Street,
		ZipCode: address.ZipCode,
	}

	return &response, nil
}

func (a *AddressUsecase) ListAddressesByUserID(ctx context.Context, userID int32) ([]dto.AddressResponse, error) {
	ctx, span := a.tracer.Start(ctx, "AddressUsecase.ListAddressesByUserID")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user_id", int(userID)),
	)

	addresses, err := a.addressRepo.ListAddressesByUserID(ctx, uint(userID), 100, 0)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	response := make([]dto.AddressResponse, len(addresses))
	for i, address := range addresses {
		response[i] = dto.AddressResponse{
			ID:      int32(address.ID),
			UserID:  int32(address.UserID),
			Country: address.Country,
			City:    address.City,
			State:   address.State,
			Street:  address.Street,
			ZipCode: address.ZipCode,
		}
	}

	return response, nil
}

func (a *AddressUsecase) UpdateAddress(ctx context.Context, req *dto.UpdateAddressRequest) error {
	ctx, span := a.tracer.Start(ctx, "AddressUsecase.UpdateAddress")
	defer span.End()

	addressToUpdate := domain.Address{
		Country: req.Country,
		City:    req.City,
		State:   req.State,
		Street:  req.Street,
		ZipCode: req.ZipCode,
	}

	updateAddressCtx, updateAddressSpan := a.tracer.Start(ctx, "addressRepo.UpdateAddress")

	_, err := a.addressRepo.UpdateAddress(updateAddressCtx, uint(req.Id), addressToUpdate)
	if err != nil {
		updateAddressSpan.RecordError(err)
		updateAddressSpan.SetStatus(codes.Error, err.Error())
		updateAddressSpan.End()

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	updateAddressSpan.End()

	return nil
}

func (a *AddressUsecase) DeleteAddress(ctx context.Context, addressID int32) error {
	ctx, span := a.tracer.Start(ctx, "AddressUsecase.DeleteAddress")
	defer span.End()

	span.SetAttributes(
		attribute.Int("address_id", int(addressID)),
	)

	err := a.addressRepo.DeleteAddress(ctx, uint(addressID))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}
