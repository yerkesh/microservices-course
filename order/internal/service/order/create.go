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

	type requestedPart struct {
		uuid     uuid.UUID
		partType model.PartType
	}

	requestedParts := []requestedPart{
		{uuid: input.HullUUID, partType: model.PartTypeHull},
		{uuid: input.EngineUUID, partType: model.PartTypeEngine},
	}
	if input.ShieldUUID != nil {
		requestedParts = append(requestedParts, requestedPart{
			uuid:     *input.ShieldUUID,
			partType: model.PartTypeShield,
		})
	}
	if input.WeaponUUID != nil {
		requestedParts = append(requestedParts, requestedPart{
			uuid:     *input.WeaponUUID,
			partType: model.PartTypeWeapon,
		})
	}

	partUUIDs := make([]string, 0, len(requestedParts))
	for _, part := range requestedParts {
		partUUIDs = append(partUUIDs, part.uuid.String())
	}

	parts, err := s.inventoryClient.ListParts(ctx, partUUIDs)
	if err != nil {
		return model.CreateOrderResult{}, err
	}

	partByUUID := make(map[uuid.UUID]model.Part, len(parts))
	for _, part := range parts {
		if part.StockQuantity <= 0 {
			return model.CreateOrderResult{}, errs.ErrOutOfStock
		}

		partByUUID[part.UUID] = part
	}

	items := make([]model.OrderItem, 0, len(requestedParts))
	var totalPrice int64
	for _, requestedPart := range requestedParts {
		part, ok := partByUUID[requestedPart.uuid]
		if !ok {
			return model.CreateOrderResult{}, errs.ErrPartNotFound
		}

		item := model.OrderItem{
			PartUUID: requestedPart.uuid,
			PartType: requestedPart.partType,
			Price:    part.Price,
		}
		items = append(items, item)
		totalPrice += item.Price
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
		Items:      items,
	}

	if err = s.txManager.Do(ctx, func(ctx context.Context) error {
		return s.orderRepo.Create(ctx, order)
	}); err != nil {
		return model.CreateOrderResult{}, err
	}

	return model.CreateOrderResult{
		OrderUUID:  orderUUID,
		TotalPrice: totalPrice,
	}, nil
}
