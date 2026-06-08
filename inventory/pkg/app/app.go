package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	apiv1 "github.com/yerkesh/inventory/internal/api/inventory/v1"
	partrepo "github.com/yerkesh/inventory/internal/repository/part"
	partsvc "github.com/yerkesh/inventory/internal/service/part"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
)

// NewInventoryAPI собирает зависимости InventoryService.
func NewInventoryAPI(txManager partsvc.TxManager, repository partsvc.Repository) inventoryv1.InventoryServiceServer {
	partService := partsvc.New(txManager, repository)

	return apiv1.New(partService)
}

// Interceptors возвращает gRPC interceptors сервиса.
func Interceptors() []grpc.ServerOption {
	return nil
}

// RegisterServices регистрирует gRPC сервисы Inventory.
func RegisterServices(server *grpc.Server, pool *pgxpool.Pool, txManager partsvc.TxManager) {
	repository := partrepo.New(pool)

	inventoryv1.RegisterInventoryServiceServer(server, NewInventoryAPI(txManager, repository))
}
