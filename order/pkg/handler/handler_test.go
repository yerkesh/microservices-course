package handler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderv1 "github.com/yerkesh/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

type fakeInventoryClient struct {
	parts map[string]*inventoryv1.Part
}

func (c fakeInventoryClient) GetPart(
	_ context.Context,
	req *inventoryv1.GetPartRequest,
	_ ...grpc.CallOption,
) (*inventoryv1.GetPartResponse, error) {
	part, ok := c.parts[req.GetUuid()]
	if !ok {
		return nil, status.Error(codes.NotFound, "part not found")
	}

	return &inventoryv1.GetPartResponse{Part: part}, nil
}

func (c fakeInventoryClient) ListParts(
	_ context.Context,
	req *inventoryv1.ListPartsRequest,
	_ ...grpc.CallOption,
) (*inventoryv1.ListPartsResponse, error) {
	parts := make([]*inventoryv1.Part, 0, len(req.GetUuids()))
	for _, partUUID := range req.GetUuids() {
		part, ok := c.parts[partUUID]
		if !ok {
			return nil, status.Error(codes.NotFound, "part not found")
		}

		parts = append(parts, part)
	}

	return &inventoryv1.ListPartsResponse{Parts: parts}, nil
}

type blockingPaymentClient struct {
	called  chan struct{}
	release chan struct{}
	once    sync.Once
	calls   atomic.Int32
}

func (c *blockingPaymentClient) PayOrder(
	ctx context.Context,
	_ *paymentv1.PayOrderRequest,
	_ ...grpc.CallOption,
) (*paymentv1.PayOrderResponse, error) {
	c.calls.Add(1)
	c.once.Do(func() {
		close(c.called)
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.release:
		return &paymentv1.PayOrderResponse{TransactionUuid: uuid.NewString()}, nil
	}
}

func newConcurrentPaymentTestHandler(paymentClient paymentv1.PaymentServiceClient) *OrderHandler {
	hullUUID := "550e8400-e29b-41d4-a716-446655440001"
	engineUUID := "550e8400-e29b-41d4-a716-446655440003"

	inventoryClient := fakeInventoryClient{
		parts: map[string]*inventoryv1.Part{
			hullUUID: {
				Uuid:          hullUUID,
				Price:         500000,
				StockQuantity: 10,
			},
			engineUUID: {
				Uuid:          engineUUID,
				Price:         300000,
				StockQuantity: 8,
			},
		},
	}

	return NewOrderHandler(inventoryClient, paymentClient, NewOrderStore())
}

func createConcurrentPaymentTestOrder(t *testing.T, h *OrderHandler) uuid.UUID {
	t.Helper()

	createRes, err := h.CreateOrder(context.Background(), &orderv1.CreateOrderRequest{
		HullUUID:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		EngineUUID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
	})
	require.NoError(t, err)

	response, ok := createRes.(*orderv1.CreateOrderResponse)
	require.True(t, ok)

	return response.OrderUUID
}

func waitForPaymentCall(t *testing.T, paymentClient *blockingPaymentClient) {
	t.Helper()

	select {
	case <-paymentClient.called:
	case <-time.After(time.Second):
		t.Fatal("payment client was not called")
	}
}

func TestPayOrderConcurrentRequestsOnlyChargeOnce(t *testing.T) {
	paymentClient := &blockingPaymentClient{
		called:  make(chan struct{}),
		release: make(chan struct{}),
	}
	h := newConcurrentPaymentTestHandler(paymentClient)
	orderUUID := createConcurrentPaymentTestOrder(t, h)

	firstPayResult := make(chan orderv1.PayOrderRes, 1)
	go func() {
		res, err := h.PayOrder(context.Background(), &orderv1.PayOrderRequest{
			PaymentMethod: orderv1.PaymentMethodCARD,
		}, orderv1.PayOrderParams{OrderUUID: orderUUID})
		require.NoError(t, err)
		firstPayResult <- res
	}()

	waitForPaymentCall(t, paymentClient)

	secondPayRes, err := h.PayOrder(context.Background(), &orderv1.PayOrderRequest{
		PaymentMethod: orderv1.PaymentMethodCARD,
	}, orderv1.PayOrderParams{OrderUUID: orderUUID})
	require.NoError(t, err)
	require.IsType(t, &orderv1.PayOrderConflict{}, secondPayRes)

	close(paymentClient.release)

	require.IsType(t, &orderv1.PayOrderResponse{}, <-firstPayResult)
	require.Equal(t, int32(1), paymentClient.calls.Load())
}

func TestCancelOrderDuringPaymentInProgressReturnsConflict(t *testing.T) {
	paymentClient := &blockingPaymentClient{
		called:  make(chan struct{}),
		release: make(chan struct{}),
	}
	h := newConcurrentPaymentTestHandler(paymentClient)
	orderUUID := createConcurrentPaymentTestOrder(t, h)

	firstPayResult := make(chan orderv1.PayOrderRes, 1)
	go func() {
		res, err := h.PayOrder(context.Background(), &orderv1.PayOrderRequest{
			PaymentMethod: orderv1.PaymentMethodCARD,
		}, orderv1.PayOrderParams{OrderUUID: orderUUID})
		require.NoError(t, err)
		firstPayResult <- res
	}()

	waitForPaymentCall(t, paymentClient)

	cancelRes, err := h.CancelOrder(context.Background(), orderv1.CancelOrderParams{OrderUUID: orderUUID})
	require.NoError(t, err)
	require.IsType(t, &orderv1.CancelOrderConflict{}, cancelRes)

	close(paymentClient.release)

	require.IsType(t, &orderv1.PayOrderResponse{}, <-firstPayResult)
	require.Equal(t, int32(1), paymentClient.calls.Load())
}
