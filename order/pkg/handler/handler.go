package handler

import (
	"context"

	apiv1 "github.com/yerkesh/order/internal/api/order/v1"
	inventoryclient "github.com/yerkesh/order/internal/client/grpc/inventory/v1"
	paymentclient "github.com/yerkesh/order/internal/client/grpc/payment/v1"
	orderrepo "github.com/yerkesh/order/internal/repository/order"
	ordersvc "github.com/yerkesh/order/internal/service/order"
	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

type noopTxManager struct{}

func (noopTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// OrderStore сохраняет совместимость с тестами на in-memory хранилище.
type OrderStore struct {
	repo ordersvc.OrderRepository
}

// OrderHandler сохраняет совместимость с кодом первой недели.
type OrderHandler = apiv1.API

// NewOrderStore создаёт новое in-memory хранилище заказов.
func NewOrderStore() *OrderStore {
	return &OrderStore{repo: orderrepo.NewMemory()}
}

// NewOrderHandler собирает OpenAPI handler заказов для тестов без PostgreSQL.
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
	orderService := ordersvc.New(noopTxManager{}, store.repo, inventoryClient, paymentClient)

	return apiv1.New(orderService)
}

// SetupServer создаёт OpenAPI сервер.
func SetupServer(h orderv1.Handler) (*orderv1.Server, error) {
	return orderv1.NewServer(h)
}
