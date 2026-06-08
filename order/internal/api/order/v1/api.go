package v1

import orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"

// API реализует OpenAPI handler OrderService.
type API struct {
	orderv1.UnimplementedHandler
	orderService OrderService
}

// New создаёт API слой заказов.
func New(orderService OrderService) *API {
	return &API{orderService: orderService}
}
