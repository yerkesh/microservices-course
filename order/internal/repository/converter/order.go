package converter

import (
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

	return model.Order{
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
	}
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
	}
}
