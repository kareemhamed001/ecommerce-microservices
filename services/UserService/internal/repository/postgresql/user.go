package postgresql

import (
	"context"
	"errors"

	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/domain"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gorm.io/gorm"
)

type UserRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db, tracer: otel.Tracer("user-repo")}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) (domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.CreateUser")
	defer span.End()
	err := gorm.G[domain.User](r.db).Create(ctx, user)

	if err != nil {
		logger.Errorf("failed to create user: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create user")
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.User{}, repository.ErrUserAlreadyExists
		}
		return domain.User{}, err
	}
	span.SetStatus(codes.Ok, "User created successfully")
	span.AddEvent("User created", trace.WithAttributes(
		attribute.Int("user.id", int(user.ID)),
		attribute.String("user.email", user.Email),
	))

	return *user, err

}

func (r *UserRepository) GetUserByID(ctx context.Context, id uint) (domain.User, error) {
	user, err := gorm.G[domain.User](r.db).
		Where("id = ?", id).
		First(ctx)
	return user, err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := gorm.G[domain.User](r.db).Where("email = ?", email).First(ctx)
	return user, err
}

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	users, err := gorm.G[domain.User](r.db).Limit(limit).Offset(offset).Find(ctx)
	return users, err
}

func (r *UserRepository) ListUsersByRole(ctx context.Context, role domain.UserRole, limit, offset int) ([]domain.User, error) {
	users, err := gorm.G[domain.User](r.db).Where("role = ?", role).Limit(limit).Offset(offset).Find(ctx)
	return users, err
}

func (r *UserRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]domain.User, error) {
	users, err := gorm.G[domain.User](r.db).
		Where("name ILIKE ? OR email ILIKE ?", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Offset(offset).
		Find(ctx)
	return users, err
}
func (r *UserRepository) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	rowsAffected, err := gorm.G[domain.User](r.db).Updates(ctx, user)
	if err != nil {
		return domain.User{}, err
	}
	if rowsAffected == 0 {
		return domain.User{}, gorm.ErrRecordNotFound
	}
	return user, err
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uint) error {
	rowsAffected, err := gorm.G[domain.User](r.db).
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
