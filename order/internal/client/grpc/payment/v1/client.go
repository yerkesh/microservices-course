package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/yerkesh/order/internal/errors"
	"github.com/yerkesh/order/internal/model"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

// Client оборачивает gRPC клиент PaymentService.
type Client struct {
	paymentClient paymentv1.PaymentServiceClient
}

// New создаёт клиент PaymentService.
func New(paymentClient paymentv1.PaymentServiceClient) *Client {
	return &Client{paymentClient: paymentClient}
}

// PayOrder оплачивает заказ через PaymentService.
func (c *Client) PayOrder(ctx context.Context, orderUUID string, method model.PaymentMethod) (uuid.UUID, error) {
	resp, err := c.paymentClient.PayOrder(ctx, &paymentv1.PayOrderRequest{
		OrderUuid:     orderUUID,
		PaymentMethod: paymentMethodToProto(method),
	})
	if err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			return uuid.Nil, errs.ErrInvalidPaymentMethod
		default:
			return uuid.Nil, fmt.Errorf("оплатить заказ: %w", err)
		}
	}

	transactionUUID, err := uuid.Parse(resp.GetTransactionUuid())
	if err != nil {
		return uuid.Nil, fmt.Errorf("разобрать UUID транзакции: %w", err)
	}

	return transactionUUID, nil
}

func paymentMethodToProto(method model.PaymentMethod) paymentv1.PaymentMethod {
	switch method {
	case model.PaymentMethodCard:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_CARD
	case model.PaymentMethodSBP:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_SBP
	case model.PaymentMethodCreditCard:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case model.PaymentMethodInvestorMoney:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY
	default:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}
