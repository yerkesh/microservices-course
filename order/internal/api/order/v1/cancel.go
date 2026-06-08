package v1

import (
	"context"

	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
)

// CancelOrder отменяет заказ.
func (a *API) CancelOrder(ctx context.Context, params orderv1.CancelOrderParams) (orderv1.CancelOrderRes, error) {
	if err := a.orderService.Cancel(ctx, params.OrderUUID); err != nil {
		return cancelErrorResponse(err), nil
	}

	return &orderv1.CancelOrderResponse{}, nil
}
