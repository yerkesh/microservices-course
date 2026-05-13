package v1

import paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"

// API реализует gRPC API PaymentService.
type API struct {
	paymentv1.UnimplementedPaymentServiceServer
	paymentService PaymentService
}

// New создаёт API слой PaymentService.
func New(paymentService PaymentService) *API {
	return &API{paymentService: paymentService}
}
