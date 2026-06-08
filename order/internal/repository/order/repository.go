package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	errs "github.com/yerkesh/order/internal/errors"
	"github.com/yerkesh/order/internal/model"
	"github.com/yerkesh/order/internal/repository/converter"
	"github.com/yerkesh/order/internal/repository/record"
)

// repository хранит заказы в PostgreSQL.
type repository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

// New создаёт PostgreSQL-репозиторий заказов.
func New(pool *pgxpool.Pool) *repository {
	return &repository{
		pool:   pool,
		getter: trmpgx.DefaultCtxGetter,
	}
}

// Create сохраняет заказ и его позиции.
func (r *repository) Create(ctx context.Context, order model.Order) error {
	const createOrderQuery = `
		INSERT INTO orders (uuid, status, created_at)
		VALUES ($1, $2, $3)
	`
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	if _, err := db.Exec(
		ctx,
		createOrderQuery,
		order.OrderUUID,
		string(order.Status),
		order.CreatedAt,
	); err != nil {
		return fmt.Errorf("создать заказ: %w", err)
	}

	if len(order.Items) == 0 {
		return nil
	}

	createItemQuery := `
		INSERT INTO order_items (order_uuid, part_uuid, part_type, price)
		VALUES ($1, $2, $3, $4)
	`
	for _, item := range order.Items {
		if _, err := db.Exec(
			ctx,
			createItemQuery,
			order.OrderUUID,
			item.PartUUID,
			string(item.PartType),
			item.Price,
		); err != nil {
			return fmt.Errorf("создать позицию заказа: %w", err)
		}
	}

	return nil
}

// Get возвращает заказ по UUID.
func (r *repository) Get(ctx context.Context, orderUUID uuid.UUID) (model.Order, error) {
	const orderQuery = `
		SELECT uuid, status, transaction_uuid::text, payment_method, created_at
		FROM orders
		WHERE uuid = $1
	`
	var (
		order           record.Order
		transactionUUID sql.NullString
		paymentMethod   sql.NullString
	)
	err := r.getter.DefaultTrOrDB(ctx, r.pool).QueryRow(ctx, orderQuery, orderUUID).Scan(
		&order.OrderUUID,
		&order.Status,
		&transactionUUID,
		&paymentMethod,
		&order.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, errs.ErrOrderNotFound
		}

		return model.Order{}, fmt.Errorf("получить заказ: %w", err)
	}
	if transactionUUID.Valid {
		parsed, parseErr := uuid.Parse(transactionUUID.String)
		if parseErr != nil {
			return model.Order{}, fmt.Errorf("разобрать UUID транзакции: %w", parseErr)
		}
		order.TransactionUUID = &parsed
	}
	if paymentMethod.Valid {
		order.PaymentMethod = &paymentMethod.String
	}

	const itemsQuery = `
		SELECT part_uuid, part_type, price
		FROM order_items
		WHERE order_uuid = $1
	`
	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, itemsQuery, orderUUID)
	if err != nil {
		return model.Order{}, fmt.Errorf("получить позиции заказа: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item record.OrderItem
		if err = rows.Scan(&item.PartUUID, &item.PartType, &item.Price); err != nil {
			return model.Order{}, fmt.Errorf("прочитать позицию заказа: %w", err)
		}

		order.Items = append(order.Items, item)
	}
	if err = rows.Err(); err != nil {
		return model.Order{}, fmt.Errorf("прочитать позиции заказа: %w", err)
	}

	return converter.OrderToModel(order), nil
}

// StartPayment атомарно переводит заказ в состояние обработки оплаты.
func (r *repository) StartPayment(ctx context.Context, orderUUID uuid.UUID) (model.Order, bool, error) {
	const query = `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE uuid = $1 AND status = $3
	`
	tag, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(
		ctx,
		query,
		orderUUID,
		model.OrderStatusPaymentProcessing,
		model.OrderStatusPendingPayment,
	)
	if err != nil {
		return model.Order{}, false, fmt.Errorf("начать оплату заказа: %w", err)
	}

	if tag.RowsAffected() == 0 {
		order, getErr := r.Get(ctx, orderUUID)
		return order, false, getErr
	}

	return model.Order{
		OrderUUID: orderUUID,
		Status:    model.OrderStatusPaymentProcessing,
	}, true, nil
}

// FinishPayment завершает оплату заказа.
func (r *repository) FinishPayment(
	ctx context.Context,
	orderUUID uuid.UUID,
	transactionUUID uuid.UUID,
	method model.PaymentMethod,
) (model.Order, bool, error) {
	const query = `
		UPDATE orders
		SET status = $2, transaction_uuid = $3, payment_method = $4, updated_at = NOW()
		WHERE uuid = $1 AND status = $5
	`
	tag, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(
		ctx,
		query,
		orderUUID,
		model.OrderStatusPaid,
		transactionUUID,
		string(method),
		model.OrderStatusPaymentProcessing,
	)
	if err != nil {
		return model.Order{}, false, fmt.Errorf("завершить оплату заказа: %w", err)
	}

	if tag.RowsAffected() == 0 {
		order, getErr := r.Get(ctx, orderUUID)
		return order, false, getErr
	}

	return model.Order{
		OrderUUID:       orderUUID,
		TransactionUUID: &transactionUUID,
		PaymentMethod:   &method,
		Status:          model.OrderStatusPaid,
	}, true, nil
}

// ResetPayment возвращает заказ в ожидание оплаты после ошибки платёжного клиента.
func (r *repository) ResetPayment(ctx context.Context, orderUUID uuid.UUID) error {
	const query = `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE uuid = $1 AND status = $3
	`
	tag, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(
		ctx,
		query,
		orderUUID,
		model.OrderStatusPendingPayment,
		model.OrderStatusPaymentProcessing,
	)
	if err != nil {
		return fmt.Errorf("сбросить оплату заказа: %w", err)
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	const existsQuery = `
		SELECT EXISTS(
			SELECT 1
			FROM orders
			WHERE uuid = $1
		)
	`
	var exists bool
	if err = r.getter.DefaultTrOrDB(ctx, r.pool).QueryRow(ctx, existsQuery, orderUUID).Scan(&exists); err != nil {
		return fmt.Errorf("проверить заказ: %w", err)
	}
	if !exists {
		return errs.ErrOrderNotFound
	}

	return nil
}

// Cancel атомарно отменяет заказ.
func (r *repository) Cancel(ctx context.Context, orderUUID uuid.UUID) (model.Order, bool, error) {
	const query = `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE uuid = $1 AND status = $3
	`
	tag, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(
		ctx,
		query,
		orderUUID,
		model.OrderStatusCancelled,
		model.OrderStatusPendingPayment,
	)
	if err != nil {
		return model.Order{}, false, fmt.Errorf("отменить заказ: %w", err)
	}

	if tag.RowsAffected() == 0 {
		order, getErr := r.Get(ctx, orderUUID)
		return order, false, getErr
	}

	return model.Order{
		OrderUUID: orderUUID,
		Status:    model.OrderStatusCancelled,
	}, true, nil
}
