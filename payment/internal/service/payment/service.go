package payment

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	errs "github.com/yerkesh/payment/internal/errors"
	"github.com/yerkesh/payment/internal/model"
)

// Service реализует бизнес-логику оплаты.
type Service struct{}

// New создаёт сервис оплаты.
func New() *Service {
	return &Service{}
}

// PayOrder проводит оплату заказа и возвращает UUID транзакции.
func (s *Service) PayOrder(_ context.Context, rawOrderUUID string, method model.PaymentMethod) (string, error) {
	orderUUID, err := uuid.Parse(rawOrderUUID)
	if err != nil || orderUUID == uuid.Nil {
		return "", errs.ErrInvalidUUID
	}

	if !isValidPaymentMethod(method) {
		return "", errs.ErrInvalidPaymentMethod
	}

	transactionUUID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	slog.Info("оплата прошла успешно",
		"order_uuid", rawOrderUUID,
		"transaction_uuid", transactionUUID.String(),
	)

	return transactionUUID.String(), nil
}

func isValidPaymentMethod(method model.PaymentMethod) bool {
	switch method {
	case model.PaymentMethodCard,
		model.PaymentMethodSBP,
		model.PaymentMethodCreditCard,
		model.PaymentMethodInvestorMoney:
		return true
	default:
		return false
	}
}
