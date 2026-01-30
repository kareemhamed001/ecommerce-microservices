package postgresql

import (
	"context"

	"github.com/kareemhamed001/e-commerce/services/UserService/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type AddressRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{db: db, tracer: otel.Tracer("address-repo")}
}

// CreateAddress(context.Context, *domain.Address) (domain.Address, error)
func (r *AddressRepository) CreateAddress(ctx context.Context, address *domain.Address) (domain.Address, error) {
	_, span := r.tracer.Start(ctx, "CreateAddress")
	defer span.End()

	err := gorm.G[domain.Address](r.db).Create(ctx, address)
	if err != nil {
		return domain.Address{}, err
	}
	return *address, nil
}

// GetAddressByID(context.Context, uint) (domain.Address, error)
func (r *AddressRepository) GetAddressByID(ctx context.Context, id uint) (domain.Address, error) {
	_, span := r.tracer.Start(ctx, "GetAddressByID")
	defer span.End()

	address, err := gorm.G[domain.Address](r.db).
		Where("id = ?", id).
		First(ctx)
	if err != nil {
		return domain.Address{}, err
	}
	return address, nil
}

// ListAddressesByUserID(context.Context, uint, int, int) ([]domain.Address, error)
func (r *AddressRepository) ListAddressesByUserID(ctx context.Context, userID uint, limit, offset int) ([]domain.Address, error) {
	_, span := r.tracer.Start(ctx, "ListAddressesByUserID")
	defer span.End()

	addresses, err := gorm.G[domain.Address](r.db).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

// UpdateAddress(context.Context, domain.Address) (domain.Address, error)
func (r *AddressRepository) UpdateAddress(ctx context.Context, address domain.Address) (domain.Address, error) {
	_, span := r.tracer.Start(ctx, "UpdateAddress")
	defer span.End()

	rowsAffected, err := gorm.G[domain.Address](r.db).Updates(ctx, address)
	if err != nil {
		return domain.Address{}, err
	}
	if rowsAffected == 0 {
		return domain.Address{}, gorm.ErrRecordNotFound
	}
	return address, nil
}

// DeleteAddress(context.Context, uint) error
func (r *AddressRepository) DeleteAddress(ctx context.Context, id uint) error {
	_, span := r.tracer.Start(ctx, "DeleteAddress")
	defer span.End()
	rowsAffected, err := gorm.G[domain.Address](r.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
