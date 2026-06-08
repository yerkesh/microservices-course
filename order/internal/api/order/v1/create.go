package v1

import (
	"context"

	"github.com/yerkesh/order/internal/model"
	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
)

// CreateOrder создаёт заказ.
func (a *API) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (orderv1.CreateOrderRes, error) {
	input := model.CreateOrderInput{
		HullUUID:   req.GetHullUUID(),
		EngineUUID: req.GetEngineUUID(),
	}
	if shieldUUID, ok := req.GetShieldUUID().Get(); ok {
		input.ShieldUUID = &shieldUUID
	}
	if weaponUUID, ok := req.GetWeaponUUID().Get(); ok {
		input.WeaponUUID = &weaponUUID
	}

	result, err := a.orderService.Create(ctx, input)
	if err != nil {
		return createErrorResponse(err), nil
	}

	return &orderv1.CreateOrderResponse{
		OrderUUID:  result.OrderUUID,
		TotalPrice: result.TotalPrice,
	}, nil
}
