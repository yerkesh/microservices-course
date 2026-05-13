package order

import (
	"context"
	"sync"

	"github.com/google/uuid"

	errs "github.com/yerkesh/order/internal/errors"
	"github.com/yerkesh/order/internal/model"
	"github.com/yerkesh/order/internal/repository/converter"
	"github.com/yerkesh/order/internal/repository/record"
)

// Repository хранит заказы в памяти.
type Repository struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]record.Order
}

// New создаёт репозиторий заказов.
func New() *Repository {
	return &Repository{orders: make(map[uuid.UUID]record.Order)}
}

// Create сохраняет новый заказ.
func (r *Repository) Create(_ context.Context, order model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.OrderUUID] = converter.OrderToRecord(order)
	return nil
}

// Get возвращает заказ по UUID.
func (r *Repository) Get(_ context.Context, orderUUID uuid.UUID) (model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, ok := r.orders[orderUUID]
	if !ok {
		return model.Order{}, errs.ErrOrderNotFound
	}

	return converter.OrderToModel(order), nil
}

// StartPayment атомарно переводит заказ в состояние обработки оплаты.
func (r *Repository) StartPayment(_ context.Context, orderUUID uuid.UUID) (model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderUUID]
	if !ok {
		return model.Order{}, errs.ErrOrderNotFound
	}

	switch model.OrderStatus(order.Status) {
	case model.OrderStatusPendingPayment:
		order.Status = string(model.OrderStatusPaymentProcessing)
		r.orders[orderUUID] = order
		return converter.OrderToModel(order), nil
	case model.OrderStatusPaymentProcessing:
		return model.Order{}, errs.ErrPaymentInProgress
	case model.OrderStatusPaid:
		return model.Order{}, errs.ErrOrderAlreadyPaid
	case model.OrderStatusCancelled:
		return model.Order{}, errs.ErrOrderCancelled
	default:
		return model.Order{}, errs.ErrOrderNotFound
	}
}

// FinishPayment завершает оплату заказа.
func (r *Repository) FinishPayment(
	_ context.Context,
	orderUUID uuid.UUID,
	transactionUUID uuid.UUID,
	method model.PaymentMethod,
) (model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderUUID]
	if !ok {
		return model.Order{}, errs.ErrOrderNotFound
	}

	order.Status = string(model.OrderStatusPaid)
	order.TransactionUUID = &transactionUUID
	paymentMethod := string(method)
	order.PaymentMethod = &paymentMethod
	r.orders[orderUUID] = order

	return converter.OrderToModel(order), nil
}

// ResetPayment возвращает заказ в ожидание оплаты после ошибки платёжного клиента.
func (r *Repository) ResetPayment(_ context.Context, orderUUID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderUUID]
	if !ok {
		return errs.ErrOrderNotFound
	}

	if model.OrderStatus(order.Status) == model.OrderStatusPaymentProcessing {
		order.Status = string(model.OrderStatusPendingPayment)
		r.orders[orderUUID] = order
	}

	return nil
}

// Cancel атомарно отменяет заказ.
func (r *Repository) Cancel(_ context.Context, orderUUID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderUUID]
	if !ok {
		return errs.ErrOrderNotFound
	}

	switch model.OrderStatus(order.Status) {
	case model.OrderStatusPendingPayment:
		order.Status = string(model.OrderStatusCancelled)
		r.orders[orderUUID] = order
		return nil
	case model.OrderStatusPaymentProcessing:
		return errs.ErrPaymentInProgress
	case model.OrderStatusPaid:
		return errs.ErrOrderAlreadyPaid
	case model.OrderStatusCancelled:
		return errs.ErrOrderCancelled
	default:
		return errs.ErrOrderNotFound
	}
}
