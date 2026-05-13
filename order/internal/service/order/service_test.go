package order

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	errs "github.com/yerkesh/order/internal/errors"
	"github.com/yerkesh/order/internal/model"
	repo "github.com/yerkesh/order/internal/repository/order"
)

type fakeInventoryClient struct {
	parts []model.Part
	err   error
}

func (c fakeInventoryClient) ListParts(_ context.Context, _ []string) ([]model.Part, error) {
	return c.parts, c.err
}

type blockingPaymentClient struct {
	called  chan struct{}
	release chan struct{}
	once    sync.Once
	calls   atomic.Int32
	err     error
}

func (c *blockingPaymentClient) PayOrder(ctx context.Context, _ string, _ model.PaymentMethod) (uuid.UUID, error) {
	c.calls.Add(1)
	c.once.Do(func() {
		close(c.called)
	})

	select {
	case <-ctx.Done():
		return uuid.Nil, ctx.Err()
	case <-c.release:
		if c.err != nil {
			return uuid.Nil, c.err
		}

		return uuid.New(), nil
	}
}

func newTestService(paymentClient PaymentClient) *Service {
	return New(repo.New(), fakeInventoryClient{
		parts: []model.Part{
			{UUID: uuid.New(), Price: 500000, StockQuantity: 10},
			{UUID: uuid.New(), Price: 300000, StockQuantity: 8},
		},
	}, paymentClient)
}

func createTestOrder(t *testing.T, service *Service) uuid.UUID {
	t.Helper()

	result, err := service.Create(context.Background(), model.CreateOrderInput{
		HullUUID:   uuid.New(),
		EngineUUID: uuid.New(),
	})
	require.NoError(t, err)

	return result.OrderUUID
}

func TestCreateSuccess(t *testing.T) {
	paymentClient := &blockingPaymentClient{called: make(chan struct{}), release: make(chan struct{})}
	service := newTestService(paymentClient)

	result, err := service.Create(context.Background(), model.CreateOrderInput{
		HullUUID:   uuid.New(),
		EngineUUID: uuid.New(),
	})

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, result.OrderUUID)
	require.Equal(t, int64(800000), result.TotalPrice)
}

func TestCreateOutOfStock(t *testing.T) {
	service := New(repo.New(), fakeInventoryClient{
		parts: []model.Part{
			{UUID: uuid.New(), Price: 500000, StockQuantity: 0},
		},
	}, &blockingPaymentClient{called: make(chan struct{}), release: make(chan struct{})})

	_, err := service.Create(context.Background(), model.CreateOrderInput{
		HullUUID:   uuid.New(),
		EngineUUID: uuid.New(),
	})

	require.ErrorIs(t, err, errs.ErrOutOfStock)
}

func TestCreateInvalidUUID(t *testing.T) {
	paymentClient := &blockingPaymentClient{called: make(chan struct{}), release: make(chan struct{})}
	service := newTestService(paymentClient)

	_, err := service.Create(context.Background(), model.CreateOrderInput{
		HullUUID:   uuid.Nil,
		EngineUUID: uuid.New(),
	})

	require.ErrorIs(t, err, errs.ErrInvalidUUID)
}

func TestPayConcurrentRequestsOnlyCallPaymentOnce(t *testing.T) {
	paymentClient := &blockingPaymentClient{called: make(chan struct{}), release: make(chan struct{})}
	service := newTestService(paymentClient)
	orderUUID := createTestOrder(t, service)

	firstResult := make(chan error, 1)
	go func() {
		_, err := service.Pay(context.Background(), orderUUID, model.PaymentMethodCard)
		firstResult <- err
	}()

	<-paymentClient.called

	_, err := service.Pay(context.Background(), orderUUID, model.PaymentMethodCard)
	require.ErrorIs(t, err, errs.ErrPaymentInProgress)

	close(paymentClient.release)
	require.NoError(t, <-firstResult)
	require.Equal(t, int32(1), paymentClient.calls.Load())
}

func TestCancelDuringPaymentReturnsConflict(t *testing.T) {
	paymentClient := &blockingPaymentClient{called: make(chan struct{}), release: make(chan struct{})}
	service := newTestService(paymentClient)
	orderUUID := createTestOrder(t, service)

	firstResult := make(chan error, 1)
	go func() {
		_, err := service.Pay(context.Background(), orderUUID, model.PaymentMethodCard)
		firstResult <- err
	}()

	<-paymentClient.called

	err := service.Cancel(context.Background(), orderUUID)
	require.ErrorIs(t, err, errs.ErrPaymentInProgress)

	close(paymentClient.release)
	require.NoError(t, <-firstResult)
}
