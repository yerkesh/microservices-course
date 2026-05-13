package tests

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yerkesh/order/tests/testutil"
)

// TestArch_TotalPrice_ComputedInService проверяет, что цена заказа равна сумме цен деталей.
func TestArch_TotalPrice_ComputedInService(t *testing.T) {
	cases := []struct {
		name string
		req  *CreateOrderRequest
		want int64
	}{
		{
			name: "только корпус и двигатель",
			req: &CreateOrderRequest{
				HullUUID:   HullAluminumUUID,
				EngineUUID: EngineIonCUUID,
			},
			want: HullAluminumPrice + EngineIonCPrice,
		},
		{
			name: "корпус двигатель и щит",
			req: &CreateOrderRequest{
				HullUUID:   HullTitaniumUUID,
				EngineUUID: EngineIonBUUID,
				ShieldUUID: testutil.Ptr(ShieldEnergyUUID),
			},
			want: HullTitaniumPrice + EngineIonBPrice + ShieldEnergyPrice,
		},
		{
			name: "все четыре слота",
			req: &CreateOrderRequest{
				HullUUID:   HullTitaniumUUID,
				EngineUUID: EngineIonBUUID,
				ShieldUUID: testutil.Ptr(ShieldEnergyUUID),
				WeaponUUID: testutil.Ptr(WeaponLaserUUID),
			},
			want: HullTitaniumPrice + EngineIonBPrice + ShieldEnergyPrice + WeaponLaserPrice,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp := createOrder(t, tc.req)
			defer resp.Body.Close()

			require.Equal(t, http.StatusCreated, resp.StatusCode)
			require.NotNil(t, result)
			assert.Equal(t, tc.want, result.TotalPrice)
		})
	}
}

func TestArch_DomainError_HullNotFound_Returns404(t *testing.T) {
	_, resp := createOrder(t, &CreateOrderRequest{
		HullUUID:   uuid.New().String(),
		EngineUUID: EngineIonCUUID,
	})
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestArch_DomainError_PayNonexistentOrder_Returns404(t *testing.T) {
	_, resp := payOrder(t, uuid.New().String(), &PayOrderRequest{PaymentMethod: "CARD"})
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestArch_DomainError_PayPaid_Returns409(t *testing.T) {
	created, createResp := createOrder(t, &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	})
	createResp.Body.Close()
	require.NotNil(t, created)

	_, payResp := payOrder(t, created.OrderUUID, &PayOrderRequest{PaymentMethod: "CARD"})
	payResp.Body.Close()

	_, resp := payOrder(t, created.OrderUUID, &PayOrderRequest{PaymentMethod: "CARD"})
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestArch_DomainError_CancelPaid_Returns409(t *testing.T) {
	created, createResp := createOrder(t, &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	})
	createResp.Body.Close()
	require.NotNil(t, created)

	_, payResp := payOrder(t, created.OrderUUID, &PayOrderRequest{PaymentMethod: "CARD"})
	payResp.Body.Close()

	_, resp := cancelOrder(t, created.OrderUUID)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestArch_DomainError_OutOfStock_Returns409(t *testing.T) {
	_, resp := createOrder(t, &CreateOrderRequest{
		HullUUID:   HullOutOfStockUUID,
		EngineUUID: EngineIonCUUID,
	})
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestArch_DomainError_CancelCancelled_Returns409(t *testing.T) {
	created, createResp := createOrder(t, &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	})
	createResp.Body.Close()
	require.NotNil(t, created)

	_, firstResp := cancelOrder(t, created.OrderUUID)
	firstResp.Body.Close()

	_, resp := cancelOrder(t, created.OrderUUID)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestArch_DomainError_InvalidPaymentMethod_Returns400(t *testing.T) {
	created, createResp := createOrder(t, &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	})
	createResp.Body.Close()
	require.NotNil(t, created)

	body := []byte(`{"payment_method": "BITCOIN"}`)
	httpReq, err := http.NewRequest(http.MethodPost,
		orderBaseURL()+"/api/v1/orders/"+created.OrderUUID+"/pay",
		bytes.NewReader(body))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestArch_DomainError_InvalidUUIDInPath_Returns400(t *testing.T) {
	resp, err := httpClient.Get(orderBaseURL() + "/api/v1/orders/not-a-uuid")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
