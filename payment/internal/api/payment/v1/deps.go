package v1

import (
	"context"

	"github.com/yerkesh/payment/internal/model"
)

// PaymentService определяет контракт бизнес-логики оплаты.
type PaymentService interface {
	PayOrder(ctx context.Context, rawOrderUUID string, method model.PaymentMethod) (string, error)
}
