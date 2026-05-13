package app

import (
	"google.golang.org/grpc"

	apiv1 "github.com/yerkesh/payment/internal/api/payment/v1"
	paymentsvc "github.com/yerkesh/payment/internal/service/payment"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

// NewPaymentAPI собирает зависимости PaymentService.
func NewPaymentAPI() paymentv1.PaymentServiceServer {
	return apiv1.New(paymentsvc.New())
}

// Interceptors возвращает gRPC interceptors сервиса.
func Interceptors() []grpc.ServerOption {
	return nil
}

// RegisterServices регистрирует gRPC сервисы Payment.
func RegisterServices(server *grpc.Server) {
	paymentv1.RegisterPaymentServiceServer(server, NewPaymentAPI())
}
