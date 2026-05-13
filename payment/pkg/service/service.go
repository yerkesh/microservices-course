package service

import (
	"context"

	"github.com/yerkesh/payment/pkg/app"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

// PaymentServer сохраняет совместимость с кодом первой недели.
type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	api paymentv1.PaymentServiceServer
}

// PayOrder обрабатывает оплату заказа.
func (s *PaymentServer) PayOrder(
	ctx context.Context,
	req *paymentv1.PayOrderRequest,
) (*paymentv1.PayOrderResponse, error) {
	if s.api == nil {
		s.api = app.NewPaymentAPI()
	}

	return s.api.PayOrder(ctx, req)
}
