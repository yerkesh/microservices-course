package converter

import (
	"github.com/google/uuid"

	"github.com/yerkesh/order/internal/model"
	"github.com/yerkesh/order/internal/repository/record"
)

// OrderToModel преобразует запись хранилища в доменную модель.
func OrderToModel(order record.Order) model.Order {
	var paymentMethod *model.PaymentMethod
	if order.PaymentMethod != nil {
		method := model.PaymentMethod(*order.PaymentMethod)
		paymentMethod = &method
	}

	result := model.Order{
		OrderUUID:       order.OrderUUID,
		HullUUID:        order.HullUUID,
		EngineUUID:      order.EngineUUID,
		ShieldUUID:      order.ShieldUUID,
		WeaponUUID:      order.WeaponUUID,
		TotalPrice:      order.TotalPrice,
		TransactionUUID: order.TransactionUUID,
		PaymentMethod:   paymentMethod,
		Status:          model.OrderStatus(order.Status),
		CreatedAt:       order.CreatedAt,
		Items:           OrderItemsToModel(order.Items),
	}

	if len(result.Items) == 0 {
		return result
	}

	result.HullUUID = uuid.Nil
	result.EngineUUID = uuid.Nil
	result.ShieldUUID = nil
	result.WeaponUUID = nil
	result.TotalPrice = 0

	for _, item := range result.Items {
		result.TotalPrice += item.Price

		switch item.PartType {
		case model.PartTypeHull:
			result.HullUUID = item.PartUUID
		case model.PartTypeEngine:
			result.EngineUUID = item.PartUUID
		case model.PartTypeShield:
			partUUID := item.PartUUID
			result.ShieldUUID = &partUUID
		case model.PartTypeWeapon:
			partUUID := item.PartUUID
			result.WeaponUUID = &partUUID
		}
	}

	return result
}

// OrderToRecord преобразует доменную модель в запись хранилища.
func OrderToRecord(order model.Order) record.Order {
	var paymentMethod *string
	if order.PaymentMethod != nil {
		method := string(*order.PaymentMethod)
		paymentMethod = &method
	}

	return record.Order{
		OrderUUID:       order.OrderUUID,
		HullUUID:        order.HullUUID,
		EngineUUID:      order.EngineUUID,
		ShieldUUID:      order.ShieldUUID,
		WeaponUUID:      order.WeaponUUID,
		TotalPrice:      order.TotalPrice,
		TransactionUUID: order.TransactionUUID,
		PaymentMethod:   paymentMethod,
		Status:          string(order.Status),
		CreatedAt:       order.CreatedAt,
		Items:           OrderItemsToRecord(order.Items),
	}
}

// OrderItemsToModel преобразует записи позиций заказа в доменную модель.
func OrderItemsToModel(items []record.OrderItem) []model.OrderItem {
	result := make([]model.OrderItem, 0, len(items))
	for _, item := range items {
		result = append(result, model.OrderItem{
			PartUUID: item.PartUUID,
			PartType: model.PartType(item.PartType),
			Price:    item.Price,
		})
	}

	return result
}

// OrderItemsToRecord преобразует доменные позиции заказа в записи хранилища.
func OrderItemsToRecord(items []model.OrderItem) []record.OrderItem {
	result := make([]record.OrderItem, 0, len(items))
	for _, item := range items {
		result = append(result, record.OrderItem{
			PartUUID: item.PartUUID,
			PartType: string(item.PartType),
			Price:    item.Price,
		})
	}

	return result
}
