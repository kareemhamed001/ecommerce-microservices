package postgresql

import (
	"context"
	"errors"

	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/domain"
	"github.com/kareemhamed001/e-commerce/services/OrderService/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

var _ domain.OrderRepository = (*OrderRepository)(nil)

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db, tracer: otel.Tracer("order-repo")}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.CreateOrder")
	defer span.End()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := gorm.G[domain.Order](tx).Create(ctx, order); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		if len(order.Items) > 0 {
			for i := range order.Items {
				order.Items[i].ID = 0
				order.Items[i].OrderID = order.ID
				if err := tx.WithContext(ctx).Omit("id").Create(&order.Items[i]).Error; err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return err
				}
			}
		}

		span.SetAttributes(attribute.Int("order.id", int(order.ID)))
		span.SetStatus(codes.Ok, "order created")
		return nil
	})
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id uint) (*domain.Order, error) {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.GetOrderByID")
	defer span.End()

	span.SetAttributes(attribute.Int("order.id", int(id)))

	var order domain.Order
	if err := r.db.WithContext(ctx).Preload("Items").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, repository.ErrOrderNotFound.Error())
			return nil, repository.ErrOrderNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "order retrieved")
	return &order, nil
}

func (r *OrderRepository) ListOrders(ctx context.Context, userID *uint, page, perPage int) ([]domain.Order, int, error) {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.ListOrders")
	defer span.End()

	query := r.db.WithContext(ctx).Model(&domain.Order{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}

	var orders []domain.Order
	if err := query.Preload("Items").Offset((page - 1) * perPage).Limit(perPage).Order("id desc").Find(&orders).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}

	span.SetAttributes(attribute.Int("orders.count", len(orders)))
	span.SetStatus(codes.Ok, "orders listed")
	return orders, int(total), nil
}

func (r *OrderRepository) AddOrderItem(ctx context.Context, item *domain.OrderItem) error {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.AddOrderItem")
	defer span.End()

	item.ID = 0
	if err := r.db.WithContext(ctx).Omit("id").Create(item).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "order item created")
	return nil
}

func (r *OrderRepository) RemoveOrderItem(ctx context.Context, orderID, itemID uint) error {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.RemoveOrderItem")
	defer span.End()

	result := r.db.WithContext(ctx).Where("id = ? AND order_id = ?", itemID, orderID).Delete(&domain.OrderItem{})
	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return result.Error
	}
	if result.RowsAffected == 0 {
		span.SetStatus(codes.Error, repository.ErrOrderItemNotFound.Error())
		return repository.ErrOrderItemNotFound
	}

	span.SetStatus(codes.Ok, "order item removed")
	return nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID uint, status domain.OrderStatus) error {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.UpdateOrderStatus")
	defer span.End()

	result := r.db.WithContext(ctx).Model(&domain.Order{}).Where("id = ?", orderID).Update("status", status)
	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return result.Error
	}
	if result.RowsAffected == 0 {
		span.SetStatus(codes.Error, repository.ErrOrderNotFound.Error())
		return repository.ErrOrderNotFound
	}

	span.SetStatus(codes.Ok, "order status updated")
	return nil
}

func (r *OrderRepository) UpdateOrderTotal(ctx context.Context, orderID uint, total float32) error {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.UpdateOrderTotal")
	defer span.End()

	result := r.db.WithContext(ctx).Model(&domain.Order{}).Where("id = ?", orderID).Update("total", total)
	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return result.Error
	}
	if result.RowsAffected == 0 {
		span.SetStatus(codes.Error, repository.ErrOrderNotFound.Error())
		return repository.ErrOrderNotFound
	}

	span.SetStatus(codes.Ok, "order total updated")
	return nil
}
