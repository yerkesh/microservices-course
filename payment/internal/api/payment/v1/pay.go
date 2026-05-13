package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/yerkesh/payment/internal/errors"
	"github.com/yerkesh/payment/internal/model"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

// PayOrder обрабатывает оплату заказа.
func (a *API) PayOrder(ctx context.Context, req *paymentv1.PayOrderRequest) (*paymentv1.PayOrderResponse, error) {
	transactionUUID, err := a.paymentService.PayOrder(
		ctx,
		req.GetOrderUuid(),
		paymentMethodToModel(req.GetPaymentMethod()),
	)
	if err != nil {
		return nil, paymentErrorToStatus(err)
	}

	return &paymentv1.PayOrderResponse{TransactionUuid: transactionUUID}, nil
}

func paymentMethodToModel(method paymentv1.PaymentMethod) model.PaymentMethod {
	switch method {
	case paymentv1.PaymentMethod_PAYMENT_METHOD_CARD:
		return model.PaymentMethodCard
	case paymentv1.PaymentMethod_PAYMENT_METHOD_SBP:
		return model.PaymentMethodSBP
	case paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD:
		return model.PaymentMethodCreditCard
	case paymentv1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY:
		return model.PaymentMethodInvestorMoney
	default:
		return model.PaymentMethodUnspecified
	}
}

func paymentErrorToStatus(err error) error {
	switch {
	case errors.Is(err, errs.ErrInvalidUUID), errors.Is(err, errs.ErrInvalidPaymentMethod):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Errorf(codes.Internal, "внутренняя ошибка: %v", err)
	}
}
