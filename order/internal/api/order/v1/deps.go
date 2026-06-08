package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/yerkesh/order/internal/model"
)

// OrderService определяет контракт бизнес-логики заказов.
type OrderService interface {
	Create(ctx context.Context, input model.CreateOrderInput) (model.CreateOrderResult, error)
	Get(ctx context.Context, orderUUID uuid.UUID) (model.Order, error)
	Pay(ctx context.Context, orderUUID uuid.UUID, method model.PaymentMethod) (uuid.UUID, error)
	Cancel(ctx context.Context, orderUUID uuid.UUID) error
}
