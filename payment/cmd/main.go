package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
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
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/yerkesh/payment/pkg/app"
)

const (
	envFile                   = "payment.env"
	dbURIEnv                  = "DB_URI"
	grpcAddress               = "127.0.0.1:50052"
	grpcMaxConnectionIdle     = 15 * time.Minute
	grpcMaxConnectionAge      = 30 * time.Minute
	grpcMaxConnectionAgeGrace = 5 * time.Second
	grpcKeepaliveTime         = 5 * time.Minute
	grpcKeepaliveTimeout      = 1 * time.Second
	grpcMinPingInterval       = 5 * time.Minute
)

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

func main() {
	if err := run(); err != nil {
		slog.Error("ошибка запуска PaymentService", "error", err)
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

	var listenConfig net.ListenConfig
	lis, err := listenConfig.Listen(context.Background(), "tcp", grpcAddress)
	if err != nil {
		return fmt.Errorf("не удалось создать listener: %w", err)
	}

	grpcServer := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     grpcMaxConnectionIdle,
		MaxConnectionAge:      grpcMaxConnectionAge,
		MaxConnectionAgeGrace: grpcMaxConnectionAgeGrace,
		Time:                  grpcKeepaliveTime,
		Timeout:               grpcKeepaliveTimeout,
	}), grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             grpcMinPingInterval,
		PermitWithoutStream: true,
	}))

	app.RegisterServices(grpcServer, pool, txManager)
	reflection.Register(grpcServer)

	slog.Info("запуск PaymentService", "адрес", grpcAddress)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			slog.Error("ошибка запуска сервера", "error", serveErr)
			cancel()
		}
	}()

	<-ctx.Done()
	slog.Info("🛑 остановка gRPC сервера")
	grpcServer.GracefulStop()
	slog.Info("✅ сервер остановлен")
	return nil
}

func newPostgresDeps(ctx context.Context, dbURI string) (*pgxpool.Pool, txManager, error) {
	pool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, nil, fmt.Errorf("создание пула соединений: %w", err)
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
