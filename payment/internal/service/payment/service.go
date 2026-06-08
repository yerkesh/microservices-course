package payment

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	errs "github.com/yerkesh/payment/internal/errors"
	"github.com/yerkesh/payment/internal/model"
)

// Service реализует бизнес-логику оплаты.
type Service struct {
	txManager TxManager
	repo      Repository
}

// New создаёт сервис оплаты.
func New(txManager TxManager, repo Repository) *Service {
	return &Service{txManager: txManager, repo: repo}
}

// PayOrder проводит оплату заказа и возвращает UUID транзакции.
func (s *Service) PayOrder(ctx context.Context, rawOrderUUID string, method model.PaymentMethod) (string, error) {
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

	payment := model.Payment{
		TransactionUUID: transactionUUID,
		OrderUUID:       orderUUID,
		PaymentMethod:   method,
		CreatedAt:       time.Now(),
	}
	if err = s.txManager.Do(ctx, func(ctx context.Context) error {
		return s.repo.Create(ctx, payment)
	}); err != nil {
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
