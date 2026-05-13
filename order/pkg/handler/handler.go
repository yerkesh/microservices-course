package handler

import (
	apiv1 "github.com/yerkesh/order/internal/api/order/v1"
	inventoryclient "github.com/yerkesh/order/internal/client/grpc/inventory/v1"
	paymentclient "github.com/yerkesh/order/internal/client/grpc/payment/v1"
	orderrepo "github.com/yerkesh/order/internal/repository/order"
	ordersvc "github.com/yerkesh/order/internal/service/order"
	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

// OrderStore сохраняет совместимость с кодом первой недели.
type OrderStore struct {
	repo *orderrepo.Repository
}

// OrderHandler сохраняет совместимость с кодом первой недели.
type OrderHandler = apiv1.API

// NewOrderStore создаёт новое хранилище заказов.
func NewOrderStore() *OrderStore {
	return &OrderStore{repo: orderrepo.New()}
}

// NewOrderHandler собирает OpenAPI handler заказов.
func NewOrderHandler(
	inventoryProtoClient inventoryv1.InventoryServiceClient,
	paymentProtoClient paymentv1.PaymentServiceClient,
	store *OrderStore,
) *OrderHandler {
	if store == nil {
		store = NewOrderStore()
	}

	inventoryClient := inventoryclient.New(inventoryProtoClient)
	paymentClient := paymentclient.New(paymentProtoClient)
	orderService := ordersvc.New(store.repo, inventoryClient, paymentClient)

	return apiv1.New(orderService)
}

// SetupServer создаёт OpenAPI сервер.
func SetupServer(h orderv1.Handler) (*orderv1.Server, error) {
	return orderv1.NewServer(h)
}
