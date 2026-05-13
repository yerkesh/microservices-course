package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/yerkesh/order/internal/errors"
	"github.com/yerkesh/order/internal/model"
)

// Pay оплачивает заказ.
func (s *Service) Pay(ctx context.Context, orderUUID uuid.UUID, method model.PaymentMethod) (uuid.UUID, error) {
	if !isValidPaymentMethod(method) {
		return uuid.Nil, errs.ErrInvalidPaymentMethod
	}

	order, err := s.orderRepo.StartPayment(ctx, orderUUID)
	if err != nil {
		return uuid.Nil, err
	}

	transactionUUID, err := s.paymentClient.PayOrder(ctx, order.OrderUUID.String(), method)
	if err != nil {
		if resetErr := s.orderRepo.ResetPayment(ctx, orderUUID); resetErr != nil {
			return uuid.Nil, errors.Join(err, resetErr)
		}

		return uuid.Nil, err
	}

	if _, err := s.orderRepo.FinishPayment(ctx, orderUUID, transactionUUID, method); err != nil {
		return uuid.Nil, err
	}

	return transactionUUID, nil
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
