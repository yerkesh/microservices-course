package v1

import (
	"errors"
	"net/http"

	errs "github.com/yerkesh/order/internal/errors"
	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
)

func createErrorResponse(err error) orderv1.CreateOrderRes {
	switch {
	case errors.Is(err, errs.ErrInvalidUUID):
		return &orderv1.CreateOrderBadRequest{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	case errors.Is(err, errs.ErrPartNotFound):
		return &orderv1.CreateOrderNotFound{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}
	case errors.Is(err, errs.ErrOutOfStock):
		return &orderv1.CreateOrderConflict{
			Code:    http.StatusConflict,
			Message: err.Error(),
		}
	default:
		return &orderv1.CreateOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "внутренняя ошибка сервера",
		}
	}
}

func payErrorResponse(err error) orderv1.PayOrderRes {
	switch {
	case errors.Is(err, errs.ErrOrderNotFound):
		return &orderv1.PayOrderNotFound{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}
	case errors.Is(err, errs.ErrInvalidPaymentMethod):
		return &orderv1.PayOrderBadRequest{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	case errors.Is(err, errs.ErrOrderAlreadyPaid),
		errors.Is(err, errs.ErrOrderCancelled),
		errors.Is(err, errs.ErrPaymentInProgress):
		return &orderv1.PayOrderConflict{
			Code:    http.StatusConflict,
			Message: err.Error(),
		}
	default:
		return &orderv1.PayOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "внутренняя ошибка сервера",
		}
	}
}

func cancelErrorResponse(err error) orderv1.CancelOrderRes {
	switch {
	case errors.Is(err, errs.ErrOrderNotFound):
		return &orderv1.CancelOrderNotFound{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}
	case errors.Is(err, errs.ErrOrderAlreadyPaid),
		errors.Is(err, errs.ErrOrderCancelled),
		errors.Is(err, errs.ErrPaymentInProgress):
		return &orderv1.CancelOrderConflict{
			Code:    http.StatusConflict,
			Message: err.Error(),
		}
	default:
		return &orderv1.CancelOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "внутренняя ошибка сервера",
		}
	}
}
