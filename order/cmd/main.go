package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/yerkesh/order/pkg/app"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/yerkesh/shared/pkg/proto/payment/v1"
)

const (
	inventoryServiceAddress = "localhost:50051"
	paymentServiceAddress   = "localhost:50052"

	// Таймауты для HTTP-сервера.
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
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
