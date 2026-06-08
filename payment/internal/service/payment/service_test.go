package payment

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	errs "github.com/yerkesh/payment/internal/errors"
	"github.com/yerkesh/payment/internal/model"
)

type fakeRepository struct {
	payments []model.Payment
	err      error
}

func (r *fakeRepository) Create(_ context.Context, payment model.Payment) error {
	if r.err != nil {
		return r.err
	}

	r.payments = append(r.payments, payment)
	return nil
}

type noopTxManager struct{}

func (noopTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestPayOrderSuccess(t *testing.T) {
	repo := &fakeRepository{}
	s := New(noopTxManager{}, repo)
	orderUUID := uuid.New()

	transactionUUID, err := s.PayOrder(context.Background(), orderUUID.String(), model.PaymentMethodCard)

	require.NoError(t, err)
	require.NotEmpty(t, transactionUUID)

	parsedTransactionUUID, err := uuid.Parse(transactionUUID)
	require.NoError(t, err)

	require.Len(t, repo.payments, 1)
	require.Equal(t, parsedTransactionUUID, repo.payments[0].TransactionUUID)
	require.Equal(t, orderUUID, repo.payments[0].OrderUUID)
	require.Equal(t, model.PaymentMethodCard, repo.payments[0].PaymentMethod)
}

func TestPayOrderInvalidUUID(t *testing.T) {
	repo := &fakeRepository{}
	s := New(noopTxManager{}, repo)

	_, err := s.PayOrder(context.Background(), "не-uuid", model.PaymentMethodCard)

	require.ErrorIs(t, err, errs.ErrInvalidUUID)
	require.Empty(t, repo.payments)
}

func TestPayOrderInvalidPaymentMethod(t *testing.T) {
	repo := &fakeRepository{}
	s := New(noopTxManager{}, repo)

	_, err := s.PayOrder(context.Background(), uuid.NewString(), model.PaymentMethod("BITCOIN"))

	require.ErrorIs(t, err, errs.ErrInvalidPaymentMethod)
	require.Empty(t, repo.payments)
}
