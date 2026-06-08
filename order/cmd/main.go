package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/yerkesh/order/internal/app"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

const (
	envFile                 = "order.env"
	dbURIEnv                = "DB_URI"
	inventoryServiceAddress = "localhost:50051"
	paymentServiceAddress   = "localhost:50052"

	// Таймауты для HTTP-сервера.
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

func main() {
	if err := run(); err != nil {
		slog.Error("ошибка запуска OrderService", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	if err := loadEnvFile(envFile); err != nil {
		return err
	}

	dbURI, err := requiredEnv(dbURIEnv)
	if err != nil {
		return err
	}

	pool, txManager, err := newPostgresDeps(ctx, dbURI)
	if err != nil {
		return err
	}
	defer pool.Close()

	slog.Info("подключение к PostgreSQL установлено")

	// Создать gRPC соединение с InventoryService
	inventoryConn, err := grpc.NewClient(inventoryServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // Интервал ping'ов для обнаружения мёртвых соединений
			Timeout:             3 * time.Second,  // Таймаут ожидания pong
			PermitWithoutStream: true,             // Держать соединение "тёплым" без активных RPC
		}))
	if err != nil {
		slog.Error("не удалось подключиться к InventoryService", "error", err)
		return err
	}
	defer inventoryConn.Close()

	// Создать gRPC клиент PaymentService
	paymentConn, err := grpc.NewClient(paymentServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // Интервал ping'ов для обнаружения мёртвых соединений
			Timeout:             3 * time.Second,  // Таймаут ожидания pong
			PermitWithoutStream: true,             // Держать соединение "тёплым" без активных RPC
		}))
	if err != nil {
		slog.Error("не удалось подключиться к PaymentService", "error", err)
		return err
	}
	defer paymentConn.Close()

	// Создать OpenAPI сервер
	orderServer, err := app.NewHTTPHandler(
		pool,
		txManager,
		inventoryv1.NewInventoryServiceClient(inventoryConn),
		paymentv1.NewPaymentServiceClient(paymentConn),
	)
	if err != nil {
		slog.Error("ошибка создания сервера OpenAPI", "error", err)
		return err
	}

	// Настроить HTTP сервер с таймаутами
	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           orderServer,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	slog.Info("запуск OrderService", "port", 8080)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	serverErrCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ошибка запуска сервера", "error", err)
			serverErrCh <- err
			cancel()
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErrCh:
		return err
	}

	slog.Info("🛑 остановка HTTP сервера")
	// Создаем контекст с таймаутом для остановки сервера
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	err = httpServer.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("❌ ошибка при остановке сервера", "error", err)
	}

	slog.Info("✅ сервер остановлен")
	return nil
}

func newPostgresDeps(ctx context.Context, dbURI string) (*pgxpool.Pool, txManager, error) {
	// DSN берём из order.env или окружения (пока хардкодим в main.go, конфиги — неделя 4)
	pool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, nil, fmt.Errorf("создание пула соединений error: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("проверка соединения с БД: %w", err)
	}

	txManager, err := manager.New(trmpgx.NewDefaultFactory(pool))
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("создание transaction manager: %w", err)
	}

	return pool, txManager, nil
}

func loadEnvFile(filename string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("получение рабочей директории: %w", err)
	}

	for {
		path := filepath.Join(dir, filename)
		if _, err = os.Stat(path); err == nil {
			return godotenv.Load(path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("проверка env-файла %s: %w", path, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("переменная окружения %s не задана", key)
	}

	return value, nil
}
