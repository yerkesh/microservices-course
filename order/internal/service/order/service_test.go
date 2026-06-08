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

func (c fakeInventoryClient) ListParts(_ context.Context, uuids []string) ([]model.Part, error) {
	if c.err != nil || c.parts != nil {
		return c.parts, c.err
	}

	parts := make([]model.Part, 0, len(uuids))
	for i, rawUUID := range uuids {
		partUUID, err := uuid.Parse(rawUUID)
		if err != nil {
			return nil, err
		}

		price := int64(300000)
		if i == 0 {
			price = 500000
		}

		parts = append(parts, model.Part{UUID: partUUID, Price: price, StockQuantity: 10})
	}

	return parts, nil
}

type blockingPaymentClient struct {
	called  chan struct{}
	release chan struct{}
	once    sync.Once
	calls   atomic.Int32
	err     error
}

type noopTxManager struct{}

func (noopTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
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
	return New(noopTxManager{}, repo.NewMemory(), fakeInventoryClient{}, paymentClient)
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
	hullUUID := uuid.New()
	engineUUID := uuid.New()
	service := New(noopTxManager{}, repo.NewMemory(), fakeInventoryClient{
		parts: []model.Part{
			{UUID: hullUUID, Price: 500000, StockQuantity: 0},
			{UUID: engineUUID, Price: 300000, StockQuantity: 8},
		},
	}, &blockingPaymentClient{called: make(chan struct{}), release: make(chan struct{})})

	_, err := service.Create(context.Background(), model.CreateOrderInput{
		HullUUID:   hullUUID,
		EngineUUID: engineUUID,
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
