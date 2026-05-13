package order

import (
	"context"

	"github.com/google/uuid"

	"github.com/yerkesh/order/internal/model"
)

// OrderRepository определяет контракт хранилища заказов.
type OrderRepository interface {
	Create(ctx context.Context, order model.Order) error
	Get(ctx context.Context, orderUUID uuid.UUID) (model.Order, error)
	StartPayment(ctx context.Context, orderUUID uuid.UUID) (model.Order, error)
	FinishPayment(ctx context.Context, orderUUID, transactionUUID uuid.UUID, method model.PaymentMethod) (model.Order, error)
	ResetPayment(ctx context.Context, orderUUID uuid.UUID) error
	Cancel(ctx context.Context, orderUUID uuid.UUID) error
}

// InventoryClient определяет контракт InventoryService.
type InventoryClient interface {
	ListParts(ctx context.Context, uuids []string) ([]model.Part, error)
}

// PaymentClient определяет контракт PaymentService.
type PaymentClient interface {
	PayOrder(ctx context.Context, orderUUID string, method model.PaymentMethod) (uuid.UUID, error)
}
