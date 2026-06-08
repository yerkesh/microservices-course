package payment

import (
	"context"

	"github.com/yerkesh/payment/internal/model"
)

// Repository определяет контракт хранилища платежей.
type Repository interface {
	Create(ctx context.Context, payment model.Payment) error
}

// TxManager определяет контракт для управления транзакциями.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
