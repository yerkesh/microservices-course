package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

// PaymentServer реализует gRPC сервис оплаты.
type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
}

// PayOrder обрабатывает оплату заказа.
func (s *PaymentServer) PayOrder(
	ctx context.Context,
	req *paymentv1.PayOrderRequest,
) (*paymentv1.PayOrderResponse, error) {
	if req.GetOrderUuid() == "" || req.GetOrderUuid() == uuid.Nil.String() {
		return nil, status.Error(codes.InvalidArgument, "invalid order uuid")
	}

	if req.GetPaymentMethod() == paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "invalid payment method")
	}

	if uuid.MustParse(req.GetOrderUuid()) == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order uuid")
	}

	transactionUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate transaction uuid: %v", err)
	}

	slog.Info("оплата прошла успешно",
		"order_uuid", req.GetOrderUuid(),
		"transaction_uuid", transactionUUID.String(),
	)

	return &paymentv1.PayOrderResponse{TransactionUuid: transactionUUID.String()}, nil
}
