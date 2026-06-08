package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	apiv1 "github.com/yerkesh/payment/internal/api/payment/v1"
	paymentrepo "github.com/yerkesh/payment/internal/repository/payment"
	paymentsvc "github.com/yerkesh/payment/internal/service/payment"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

// NewPaymentAPI собирает зависимости PaymentService.
func NewPaymentAPI(txManager paymentsvc.TxManager, repository paymentsvc.Repository) paymentv1.PaymentServiceServer {
	paymentService := paymentsvc.New(txManager, repository)

	return apiv1.New(paymentService)
}

// Interceptors возвращает gRPC interceptors сервиса.
func Interceptors() []grpc.ServerOption {
	return nil
}

// RegisterServices регистрирует gRPC сервисы Payment.
func RegisterServices(server *grpc.Server, pool *pgxpool.Pool, txManager paymentsvc.TxManager) {
	repository := paymentrepo.New(pool)

	paymentv1.RegisterPaymentServiceServer(server, NewPaymentAPI(txManager, repository))
}
