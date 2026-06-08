package v1

import (
	"context"

	"github.com/yerkesh/order/internal/converter"
	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
)

// PayOrder оплачивает заказ.
func (a *API) PayOrder(ctx context.Context, req *orderv1.PayOrderRequest, params orderv1.PayOrderParams) (orderv1.PayOrderRes, error) {
	transactionUUID, err := a.orderService.Pay(
		ctx,
		params.OrderUUID,
		converter.PaymentMethodToModel(req.GetPaymentMethod()),
	)
	if err != nil {
		return payErrorResponse(err), nil
	}

	return &orderv1.PayOrderResponse{TransactionUUID: transactionUUID}, nil
}
