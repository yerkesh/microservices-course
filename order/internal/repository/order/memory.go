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

// MemoryRepository хранит заказы в памяти для быстрых unit-тестов.
type MemoryRepository struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]record.Order
}

// NewMemory создаёт in-memory репозиторий заказов.
func NewMemory() *MemoryRepository {
	return &MemoryRepository{orders: make(map[uuid.UUID]record.Order)}
}

// Create сохраняет новый заказ.
func (r *MemoryRepository) Create(_ context.Context, order model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.OrderUUID] = converter.OrderToRecord(order)
	return nil
}

// Get возвращает заказ по UUID.
func (r *MemoryRepository) Get(_ context.Context, orderUUID uuid.UUID) (model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, ok := r.orders[orderUUID]
	if !ok {
		return model.Order{}, errs.ErrOrderNotFound
	}

	return converter.OrderToModel(order), nil
}

// StartPayment атомарно переводит заказ в состояние обработки оплаты.
func (r *MemoryRepository) StartPayment(_ context.Context, orderUUID uuid.UUID) (model.Order, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderUUID]
	if !ok {
		return model.Order{}, false, errs.ErrOrderNotFound
	}

	if model.OrderStatus(order.Status) != model.OrderStatusPendingPayment {
		return converter.OrderToModel(order), false, nil
	}

	order.Status = string(model.OrderStatusPaymentProcessing)
	r.orders[orderUUID] = order

	return converter.OrderToModel(order), true, nil
}

// FinishPayment завершает оплату заказа.
func (r *MemoryRepository) FinishPayment(
	_ context.Context,
	orderUUID uuid.UUID,
	transactionUUID uuid.UUID,
	method model.PaymentMethod,
) (model.Order, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderUUID]
	if !ok {
		return model.Order{}, false, errs.ErrOrderNotFound
	}
	if model.OrderStatus(order.Status) != model.OrderStatusPaymentProcessing {
		return converter.OrderToModel(order), false, nil
	}

	order.Status = string(model.OrderStatusPaid)
	order.TransactionUUID = &transactionUUID
	paymentMethod := string(method)
	order.PaymentMethod = &paymentMethod
	r.orders[orderUUID] = order

	return converter.OrderToModel(order), true, nil
}

// ResetPayment возвращает заказ в ожидание оплаты после ошибки платёжного клиента.
func (r *MemoryRepository) ResetPayment(_ context.Context, orderUUID uuid.UUID) error {
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
func (r *MemoryRepository) Cancel(_ context.Context, orderUUID uuid.UUID) (model.Order, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderUUID]
	if !ok {
		return model.Order{}, false, errs.ErrOrderNotFound
	}

	if model.OrderStatus(order.Status) != model.OrderStatusPendingPayment {
		return converter.OrderToModel(order), false, nil
	}

	order.Status = string(model.OrderStatusCancelled)
	r.orders[orderUUID] = order

	return converter.OrderToModel(order), true, nil
}
