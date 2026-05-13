package payment

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	errs "github.com/yerkesh/payment/internal/errors"
	"github.com/yerkesh/payment/internal/model"
)

func TestPayOrderSuccess(t *testing.T) {
	s := New()

	transactionUUID, err := s.PayOrder(context.Background(), uuid.NewString(), model.PaymentMethodCard)

	require.NoError(t, err)
	require.NotEmpty(t, transactionUUID)
	_, err = uuid.Parse(transactionUUID)
	require.NoError(t, err)
}

func TestPayOrderInvalidUUID(t *testing.T) {
	s := New()

	_, err := s.PayOrder(context.Background(), "не-uuid", model.PaymentMethodCard)

	require.ErrorIs(t, err, errs.ErrInvalidUUID)
}

func TestPayOrderInvalidPaymentMethod(t *testing.T) {
	s := New()

	_, err := s.PayOrder(context.Background(), uuid.NewString(), model.PaymentMethod("BITCOIN"))

	require.ErrorIs(t, err, errs.ErrInvalidPaymentMethod)
}
