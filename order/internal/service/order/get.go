package order

import (
	"context"

	"github.com/google/uuid"

	"github.com/yerkesh/order/internal/model"
)

// Get возвращает заказ по UUID.
func (s *Service) Get(ctx context.Context, orderUUID uuid.UUID) (model.Order, error) {
	return s.orderRepo.Get(ctx, orderUUID)
}
