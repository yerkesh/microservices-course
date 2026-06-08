package converter

import (
	"github.com/yerkesh/order/internal/model"
	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
)

// OrderToDTO преобразует доменный заказ в DTO OpenAPI.
func OrderToDTO(order model.Order) *orderv1.OrderDto {
	var shieldUUID orderv1.OptNilUUID
	if order.ShieldUUID != nil {
		shieldUUID = orderv1.NewOptNilUUID(*order.ShieldUUID)
	}

	var weaponUUID orderv1.OptNilUUID
	if order.WeaponUUID != nil {
		weaponUUID = orderv1.NewOptNilUUID(*order.WeaponUUID)
	}

	var transactionUUID orderv1.OptNilUUID
	if order.TransactionUUID != nil {
		transactionUUID = orderv1.NewOptNilUUID(*order.TransactionUUID)
	}

	var paymentMethod orderv1.OptNilPaymentMethod
	if order.PaymentMethod != nil {
		paymentMethod = orderv1.NewOptNilPaymentMethod(PaymentMethodToDTO(*order.PaymentMethod))
	}

	return &orderv1.OrderDto{
		OrderUUID:       order.OrderUUID,
		HullUUID:        order.HullUUID,
		EngineUUID:      order.EngineUUID,
		ShieldUUID:      shieldUUID,
		WeaponUUID:      weaponUUID,
		TotalPrice:      order.TotalPrice,
		TransactionUUID: transactionUUID,
		PaymentMethod:   paymentMethod,
		Status:          OrderStatusToDTO(order.Status),
		CreatedAt:       order.CreatedAt,
	}
}

// PaymentMethodToModel преобразует OpenAPI enum в доменную модель.
func PaymentMethodToModel(method orderv1.PaymentMethod) model.PaymentMethod {
	switch method {
	case orderv1.PaymentMethodCARD:
		return model.PaymentMethodCard
	case orderv1.PaymentMethodSBP:
		return model.PaymentMethodSBP
	case orderv1.PaymentMethodCREDITCARD:
		return model.PaymentMethodCreditCard
	case orderv1.PaymentMethodINVESTORMONEY:
		return model.PaymentMethodInvestorMoney
	default:
		return ""
	}
}

// PaymentMethodToDTO преобразует доменный enum в OpenAPI enum.
func PaymentMethodToDTO(method model.PaymentMethod) orderv1.PaymentMethod {
	switch method {
	case model.PaymentMethodCard:
		return orderv1.PaymentMethodCARD
	case model.PaymentMethodSBP:
		return orderv1.PaymentMethodSBP
	case model.PaymentMethodCreditCard:
		return orderv1.PaymentMethodCREDITCARD
	case model.PaymentMethodInvestorMoney:
		return orderv1.PaymentMethodINVESTORMONEY
	default:
		return ""
	}
}

// OrderStatusToDTO преобразует внутренний статус заказа в публичный enum.
func OrderStatusToDTO(status model.OrderStatus) orderv1.OrderStatus {
	switch status {
	case model.OrderStatusPaid:
		return orderv1.OrderStatusPAID
	case model.OrderStatusCancelled:
		return orderv1.OrderStatusCANCELLED
	default:
		return orderv1.OrderStatusPENDINGPAYMENT
	}
}
