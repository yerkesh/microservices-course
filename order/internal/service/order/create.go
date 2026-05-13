package order

import (
	"context"
	"time"

	"github.com/google/uuid"

	errs "github.com/yerkesh/order/internal/errors"
	"github.com/yerkesh/order/internal/model"
)

// Create создаёт заказ.
func (s *Service) Create(ctx context.Context, input model.CreateOrderInput) (model.CreateOrderResult, error) {
	if input.HullUUID == uuid.Nil || input.EngineUUID == uuid.Nil {
		return model.CreateOrderResult{}, errs.ErrInvalidUUID
	}
	if input.ShieldUUID != nil && *input.ShieldUUID == uuid.Nil {
		return model.CreateOrderResult{}, errs.ErrInvalidUUID
	}
	if input.WeaponUUID != nil && *input.WeaponUUID == uuid.Nil {
		return model.CreateOrderResult{}, errs.ErrInvalidUUID
	}

	partUUIDs := []string{
		input.HullUUID.String(),
		input.EngineUUID.String(),
	}
	if input.ShieldUUID != nil {
		partUUIDs = append(partUUIDs, input.ShieldUUID.String())
	}
	if input.WeaponUUID != nil {
		partUUIDs = append(partUUIDs, input.WeaponUUID.String())
	}

	parts, err := s.inventoryClient.ListParts(ctx, partUUIDs)
	if err != nil {
		return model.CreateOrderResult{}, err
	}

	var totalPrice int64
	for _, part := range parts {
		if part.StockQuantity <= 0 {
			return model.CreateOrderResult{}, errs.ErrOutOfStock
		}
		totalPrice += part.Price
	}

	orderUUID, err := uuid.NewRandom()
	if err != nil {
		return model.CreateOrderResult{}, err
	}

	order := model.Order{
		OrderUUID:  orderUUID,
		HullUUID:   input.HullUUID,
		EngineUUID: input.EngineUUID,
		ShieldUUID: input.ShieldUUID,
		WeaponUUID: input.WeaponUUID,
		TotalPrice: totalPrice,
		Status:     model.OrderStatusPendingPayment,
		CreatedAt:  time.Now(),
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return model.CreateOrderResult{}, err
	}

	return model.CreateOrderResult{
		OrderUUID:  orderUUID,
		TotalPrice: totalPrice,
	}, nil
}
