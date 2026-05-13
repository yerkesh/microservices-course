package app

import (
	"net/http"

	apiv1 "github.com/yerkesh/order/internal/api/order/v1"
	inventoryclient "github.com/yerkesh/order/internal/client/grpc/inventory/v1"
	paymentclient "github.com/yerkesh/order/internal/client/grpc/payment/v1"
	orderrepo "github.com/yerkesh/order/internal/repository/order"
	ordersvc "github.com/yerkesh/order/internal/service/order"
	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

// NewHTTPHandler собирает HTTP handler OrderService.
func NewHTTPHandler(
	inventoryProtoClient inventoryv1.InventoryServiceClient,
	paymentProtoClient paymentv1.PaymentServiceClient,
) (http.Handler, error) {
	orderRepo := orderrepo.New()
	inventoryClient := inventoryclient.New(inventoryProtoClient)
	paymentClient := paymentclient.New(paymentProtoClient)
	orderService := ordersvc.New(orderRepo, inventoryClient, paymentClient)
	orderAPI := apiv1.New(orderService)

	return orderv1.NewServer(orderAPI)
}
