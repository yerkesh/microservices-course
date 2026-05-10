package handler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

// Order представляет заказ на постройку космического корабля.
type Order struct {
	OrderUUID       uuid.UUID
	HullUUID        uuid.UUID
	EngineUUID      uuid.UUID
	ShieldUUID      *uuid.UUID // опциональный
	WeaponUUID      *uuid.UUID // опциональный
	TotalPrice      int64      // в копейках
	TransactionUUID *uuid.UUID
	PaymentMethod   *string
	Status          string // PENDING_PAYMENT, PAID, CANCELLED
	CreatedAt       time.Time
}

// OrderStore — хранилище заказов (in-memory).
type OrderStore struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]Order
}

// NewOrderStore создаёт новое пустое хранилище заказов.
func NewOrderStore() *OrderStore {
	return &OrderStore{
		orders: make(map[uuid.UUID]Order),
	}
}

// OrderHandler реализует интерфейс orderv1.Handler, сгенерированный ogen.
type OrderHandler struct {
	orderv1.UnimplementedHandler
	inventoryClient inventoryv1.InventoryServiceClient
	paymentClient   paymentv1.PaymentServiceClient
	store           *OrderStore
}

// NewOrderHandler создаёт новый обработчик заказов.
func NewOrderHandler(
	inventoryClient inventoryv1.InventoryServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
	store *OrderStore,
) *OrderHandler {
	return &OrderHandler{
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
		store:           store,
	}
}

// SetupServer создаёт OpenAPI сервер на основе обработчика.
func SetupServer(h *OrderHandler) (*orderv1.Server, error) {
	return orderv1.NewServer(h)
}

// GetOrder реализует операцию getOrder (пример реализации).
// GET /api/v1/orders/{order_uuid}.
func (h *OrderHandler) GetOrder(_ context.Context, params orderv1.GetOrderParams) (orderv1.GetOrderRes, error) {
	// 1. Найти заказ в store (с блокировкой для thread-safety)
	h.store.mu.RLock()
	order, ok := h.store.orders[params.OrderUUID]
	h.store.mu.RUnlock()

	// 2. Если не найден — вернуть 404
	if !ok {
		return &orderv1.GetOrderNotFound{
			Code:    http.StatusNotFound,
			Message: "заказ не найден",
		}, nil
	}

	// 3. Преобразовать в DTO и вернуть
	var shieldUUID orderv1.OptNilUUID
	if order.ShieldUUID != nil {
		shieldUUID = orderv1.NewOptNilUUID(*order.ShieldUUID)
	}

	var weaponUUID orderv1.OptNilUUID
	if order.WeaponUUID != nil {
		weaponUUID = orderv1.NewOptNilUUID(*order.WeaponUUID)
	}

	var transactionUUID orderv1.OptNilUUID
	if order.TransactionUUID != nil {
		transactionUUID = orderv1.NewOptNilUUID(*order.TransactionUUID)
	}

	var paymentMethod orderv1.OptNilPaymentMethod
	if order.PaymentMethod != nil {
		paymentMethod = orderv1.NewOptNilPaymentMethod(orderv1.PaymentMethod(*order.PaymentMethod))
	}

	return &orderv1.OrderDto{
		OrderUUID:       order.OrderUUID,
		HullUUID:        order.HullUUID,
		EngineUUID:      order.EngineUUID,
		ShieldUUID:      shieldUUID,
		WeaponUUID:      weaponUUID,
		TotalPrice:      order.TotalPrice,
		TransactionUUID: transactionUUID,
		PaymentMethod:   paymentMethod,
		Status:          orderv1.OrderStatus(order.Status),
		CreatedAt:       order.CreatedAt,
	}, nil
}

// CreateOrder реализует операцию createOrder
// POST /api/v1/orders.
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (orderv1.CreateOrderRes, error) {
	if req.GetHullUUID() == uuid.Nil || req.GetEngineUUID() == uuid.Nil {
		return &orderv1.CreateOrderBadRequest{
			Code:    http.StatusBadRequest,
			Message: "hull_uuid и engine_uuid обязательны",
		}, nil
	}

	part, err := h.inventoryClient.GetPart(ctx, &inventoryv1.GetPartRequest{
		Uuid: req.GetHullUUID().String(),
	})
	if err != nil {
		return &orderv1.CreateOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка при получении детали",
		}, nil
	}

	if part.GetPart().GetStockQuantity() == 0 {
		return &orderv1.CreateOrderConflict{
			Code:    http.StatusConflict,
			Message: "нет в наличии",
		}, nil
	}

	totalPrice := part.GetPart().GetPrice() * part.GetPart().GetStockQuantity()
	orderUUID, err := uuid.NewRandom()
	if err != nil {
		return &orderv1.CreateOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка при генерации UUID",
		}, nil
	}

	h.store.mu.Lock()
	h.store.orders[orderUUID] = Order{
		OrderUUID:       orderUUID,
		HullUUID:        req.GetHullUUID(),
		EngineUUID:      req.GetEngineUUID(),
		ShieldUUID:      nil,
		WeaponUUID:      nil,
		TotalPrice:      totalPrice,
		TransactionUUID: nil,
		PaymentMethod:   nil,
		Status:          "PENDING_PAYMENT",
		CreatedAt:       time.Now(),
	}
	h.store.mu.Unlock()

	return &orderv1.CreateOrderResponse{
		OrderUUID:  orderUUID,
		TotalPrice: totalPrice,
	}, nil
}

// PayOrder реализует операцию payOrder.
// POST /api/v1/orders/{order_uuid}/pay.
func (h *OrderHandler) PayOrder(ctx context.Context, req *orderv1.PayOrderRequest, params orderv1.PayOrderParams) (orderv1.PayOrderRes, error) {
	// 1. Найти заказ в store
	// 2. Проверить статус == PENDING_PAYMENT
	// 3. Вызвать h.paymentClient.PayOrder для обработки платежа
	// 4. Обновить статус на PAID и сохранить transaction_uuid
	// 5. Вернуть transaction_uuid
	h.store.mu.RLock()
	order, ok := h.store.orders[params.OrderUUID]
	h.store.mu.RUnlock()
	if !ok || order.Status != "PENDING_PAYMENT" {
		return &orderv1.PayOrderBadRequest{
			Code:    http.StatusBadRequest,
			Message: "заказ не найден или не в ожидании оплаты",
		}, nil
	}

	rawMethod := req.GetPaymentMethod()

	val, ok := paymentv1.PaymentMethod_value[string(rawMethod)]
	if !ok {
		return nil, fmt.Errorf("unknown payment method: %s", rawMethod)
	}

	paymentMethod := paymentv1.PaymentMethod(val)

	transactionUUID, err := h.paymentClient.PayOrder(ctx, &paymentv1.PayOrderRequest{
		OrderUuid:     params.OrderUUID.String(),
		PaymentMethod: paymentMethod,
	})
	if err != nil {
		return &orderv1.PayOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка при обработке платежа",
		}, nil
	}

	transactionUUIDParsed, err := uuid.Parse(transactionUUID.GetTransactionUuid())
	if err != nil {
		return &orderv1.PayOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка при парсинге UUID транзакции",
		}, nil
	}

	h.store.mu.Lock()
	order.Status = "PAID"
	order.TransactionUUID = &transactionUUIDParsed
	h.store.orders[params.OrderUUID] = order
	h.store.mu.Unlock()

	return &orderv1.PayOrderResponse{
		TransactionUUID: transactionUUIDParsed,
	}, nil
}

// CancelOrder реализует операцию cancelOrder.
// POST /api/v1/orders/{order_uuid}/cancel.
func (h *OrderHandler) CancelOrder(ctx context.Context, params orderv1.CancelOrderParams) (orderv1.CancelOrderRes, error) {
	h.store.mu.RLock()
	order, ok := h.store.orders[params.OrderUUID]
	h.store.mu.RUnlock()
	if !ok || order.Status != "PENDING_PAYMENT" {
		return &orderv1.CancelOrderBadRequest{
			Code:    http.StatusBadRequest,
			Message: "заказ не найден или не в ожидании оплаты",
		}, nil
	}

	h.store.mu.Lock()
	order.Status = "CANCELLED"
	h.store.orders[params.OrderUUID] = order
	h.store.mu.Unlock()

	return &orderv1.CancelOrderResponse{}, nil
}
