package v1

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/yerkesh/payment/internal/errors"
	"github.com/yerkesh/payment/internal/model"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

type fakePaymentService struct {
	transactionUUID string
	err             error
}

func (s fakePaymentService) PayOrder(_ context.Context, _ string, _ model.PaymentMethod) (string, error) {
	return s.transactionUUID, s.err
}

func TestPayOrderSuccess(t *testing.T) {
	transactionUUID := uuid.NewString()
	api := New(fakePaymentService{transactionUUID: transactionUUID})

	resp, err := api.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     uuid.NewString(),
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
	})

	require.NoError(t, err)
	require.Equal(t, transactionUUID, resp.GetTransactionUuid())
}

func TestPayOrderInvalidMethod(t *testing.T) {
	api := New(fakePaymentService{err: errs.ErrInvalidPaymentMethod})

	_, err := api.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     uuid.NewString(),
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED,
	})

	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
