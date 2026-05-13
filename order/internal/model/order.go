package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentMethod описывает способ оплаты.
type PaymentMethod string

const (
	PaymentMethodCard          PaymentMethod = "CARD"
	PaymentMethodSBP           PaymentMethod = "SBP"
	PaymentMethodCreditCard    PaymentMethod = "CREDIT_CARD"
	PaymentMethodInvestorMoney PaymentMethod = "INVESTOR_MONEY"
)

// OrderStatus описывает состояние заказа.
type OrderStatus string

const (
	OrderStatusPendingPayment    OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaymentProcessing OrderStatus = "PAYMENT_PROCESSING"
	OrderStatusPaid              OrderStatus = "PAID"
	OrderStatusCancelled         OrderStatus = "CANCELLED"
)

// Part описывает деталь в доменной модели OrderService.
type Part struct {
	UUID          uuid.UUID
	Price         int64
	StockQuantity int64
}

// CreateOrderInput описывает входные данные создания заказа.
type CreateOrderInput struct {
	HullUUID   uuid.UUID
	EngineUUID uuid.UUID
	ShieldUUID *uuid.UUID
	WeaponUUID *uuid.UUID
}

// CreateOrderResult описывает результат создания заказа.
type CreateOrderResult struct {
	OrderUUID  uuid.UUID
	TotalPrice int64
}

// Order описывает заказ на постройку корабля.
type Order struct {
	OrderUUID       uuid.UUID
	HullUUID        uuid.UUID
	EngineUUID      uuid.UUID
	ShieldUUID      *uuid.UUID
	WeaponUUID      *uuid.UUID
	TotalPrice      int64
	TransactionUUID *uuid.UUID
	PaymentMethod   *PaymentMethod
	Status          OrderStatus
	CreatedAt       time.Time
}
