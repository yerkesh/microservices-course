package payment

import (
	"context"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yerkesh/payment/internal/model"
)

// repository хранит платежи в PostgreSQL.
type repository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

// New создаёт PostgreSQL-репозиторий платежей.
func New(pool *pgxpool.Pool) *repository {
	return &repository{
		pool:   pool,
		getter: trmpgx.DefaultCtxGetter,
	}
}

// Create сохраняет платёж.
func (r *repository) Create(ctx context.Context, payment model.Payment) error {
	const query = `
		INSERT INTO payments (transaction_uuid, order_uuid, payment_method, created_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(
		ctx,
		query,
		payment.TransactionUUID,
		payment.OrderUUID,
		string(payment.PaymentMethod),
		payment.CreatedAt,
	); err != nil {
		return fmt.Errorf("создать платёж: %w", err)
	}

	return nil
}
