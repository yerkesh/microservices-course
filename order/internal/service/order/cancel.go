package order

import (
	"context"

	"github.com/google/uuid"

	errs "github.com/yerkesh/order/internal/errors"
	"github.com/yerkesh/order/internal/model"
)

// Cancel отменяет заказ.
func (s *Service) Cancel(ctx context.Context, orderUUID uuid.UUID) error {
	order, updated, err := s.orderRepo.Cancel(ctx, orderUUID)
	if err != nil {
		return err
	}
	if updated {
		return nil
	}

	switch order.Status {
	case model.OrderStatusPaymentProcessing:
		return errs.ErrPaymentInProgress
	case model.OrderStatusPaid:
		return errs.ErrOrderAlreadyPaid
	case model.OrderStatusCancelled:
		return errs.ErrOrderCancelled
	default:
		return errs.ErrOrderNotFound
	}
}
