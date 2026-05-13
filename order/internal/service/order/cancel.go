package order

import (
	"context"

	"github.com/google/uuid"
)

// Cancel отменяет заказ.
func (s *Service) Cancel(ctx context.Context, orderUUID uuid.UUID) error {
	return s.orderRepo.Cancel(ctx, orderUUID)
}
