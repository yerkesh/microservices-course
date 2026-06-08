package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/yerkesh/order/internal/converter"
	errs "github.com/yerkesh/order/internal/errors"
	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
)

// GetOrder возвращает заказ.
func (a *API) GetOrder(ctx context.Context, params orderv1.GetOrderParams) (orderv1.GetOrderRes, error) {
	order, err := a.orderService.Get(ctx, params.OrderUUID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return &orderv1.GetOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		}

		return &orderv1.GetOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "внутренняя ошибка сервера",
		}, nil
	}

	return converter.OrderToDTO(order), nil
}
