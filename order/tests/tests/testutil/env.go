package testutil

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	invApp "github.com/yerkesh/inventory/pkg/app"
	"github.com/yerkesh/order/internal/app"
	payApp "github.com/yerkesh/payment/pkg/app"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

const bufSize = 1024 * 1024

// Env — изолированное тестовое окружение: свои БД, свои сервисы.
// Каждый параллельный тест получает свой Env и не пересекается с другими.
type Env struct {
	HTTPClient *http.Client
	BaseURL    string

	InventoryClient inventoryv1.InventoryServiceClient
	PaymentClient   paymentv1.PaymentServiceClient

	// Пулы прямого доступа к БД для проверок состояния и seed-данных.
	OrderPool     *pgxpool.Pool
	InventoryPool *pgxpool.Pool
	PaymentPool   *pgxpool.Pool

	// Имена изолированных БД (полезно для отладки).
	OrderDBName     string
	InventoryDBName string
	PaymentDBName   string
}

// NewEnv поднимает окружение для одного теста и регистрирует cleanup.
func NewEnv(t *testing.T) *Env {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	orderDB := createIsolatedDB(ctx, t, "order", "../../../migrations/order")
	t.Cleanup(orderDB.cleanup)

	inventoryDB := createIsolatedDB(ctx, t, "inventory", "../../../migrations/inventory")
	t.Cleanup(inventoryDB.cleanup)

	paymentDB := createIsolatedDB(ctx, t, "payment", "../../../migrations/payment")
	t.Cleanup(paymentDB.cleanup)

	orderPool := newPool(ctx, t, "orderPool", orderDB.DSN)
	inventoryPool := newPool(ctx, t, "inventoryPool", inventoryDB.DSN)
	paymentPool := newPool(ctx, t, "paymentPool", paymentDB.DSN)

	orderTxManager := newTxManager(t, "orderTxManager", orderPool)
	inventoryTxManager := newTxManager(t, "inventoryTxManager", inventoryPool)
	paymentTxManager := newTxManager(t, "paymentTxManager", paymentPool)

	// Inventory gRPC через bufconn.
	invLis := bufconn.Listen(bufSize)
	invServer := grpc.NewServer(invApp.Interceptors()...)
	invApp.RegisterServices(invServer, inventoryPool, inventoryTxManager)
	go func() {
		if serveErr := invServer.Serve(invLis); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			t.Errorf("inventory server: %v", serveErr)
		}
	}()
	t.Cleanup(invServer.Stop)

	invConn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return invLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("invConn: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := invConn.Close(); closeErr != nil {
			t.Errorf("close inventory connection: %v", closeErr)
		}
	})
	invClient := inventoryv1.NewInventoryServiceClient(invConn)

	// Payment gRPC через bufconn.
	payLis := bufconn.Listen(bufSize)
	payServer := grpc.NewServer(payApp.Interceptors()...)
	payApp.RegisterServices(payServer, paymentPool, paymentTxManager)
	go func() {
		if serveErr := payServer.Serve(payLis); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			t.Errorf("payment server: %v", serveErr)
		}
	}()
	t.Cleanup(payServer.Stop)

	payConn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return payLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("payConn: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := payConn.Close(); closeErr != nil {
			t.Errorf("close payment connection: %v", closeErr)
		}
	})
	payClient := paymentv1.NewPaymentServiceClient(payConn)

	// Order HTTP через httptest.
	orderHandler, err := app.NewHTTPHandler(orderPool, orderTxManager, invClient, payClient)
	if err != nil {
		t.Fatalf("order handler: %v", err)
	}
	ts := httptest.NewServer(orderHandler)
	t.Cleanup(ts.Close)

	return &Env{
		HTTPClient:      &http.Client{Timeout: 10 * time.Second},
		BaseURL:         ts.URL,
		InventoryClient: invClient,
		PaymentClient:   payClient,
		OrderPool:       orderPool,
		InventoryPool:   inventoryPool,
		PaymentPool:     paymentPool,
		OrderDBName:     orderDB.Name,
		InventoryDBName: inventoryDB.Name,
		PaymentDBName:   paymentDB.Name,
	}
}

func newPool(ctx context.Context, t *testing.T, name, dsn string) *pgxpool.Pool {
	t.Helper()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	t.Cleanup(pool.Close)

	return pool
}

func newTxManager(t *testing.T, name string, pool *pgxpool.Pool) txManager {
	t.Helper()

	txManager, err := manager.New(trmpgx.NewDefaultFactory(pool))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}

	return txManager
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
