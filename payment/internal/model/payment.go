package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentMethod описывает способ оплаты.
type PaymentMethod string

const (
	PaymentMethodUnspecified   PaymentMethod = "UNSPECIFIED"
	PaymentMethodCard          PaymentMethod = "CARD"
	PaymentMethodSBP           PaymentMethod = "SBP"
	PaymentMethodCreditCard    PaymentMethod = "CREDIT_CARD"
	PaymentMethodInvestorMoney PaymentMethod = "INVESTOR_MONEY"
)

// Payment описывает сохранённый платёж.
type Payment struct {
	TransactionUUID uuid.UUID
	OrderUUID       uuid.UUID
	PaymentMethod   PaymentMethod
	CreatedAt       time.Time
}
