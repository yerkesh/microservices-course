package app

import (
	"google.golang.org/grpc"

	apiv1 "github.com/yerkesh/inventory/internal/api/inventory/v1"
	repo "github.com/yerkesh/inventory/internal/repository/part"
	partsvc "github.com/yerkesh/inventory/internal/service/part"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
)

// NewInventoryAPI собирает зависимости InventoryService.
func NewInventoryAPI() inventoryv1.InventoryServiceServer {
	partRepo := repo.New()
	partService := partsvc.New(partRepo)

	return apiv1.New(partService)
}

// Interceptors возвращает gRPC interceptors сервиса.
func Interceptors() []grpc.ServerOption {
	return nil
}

// RegisterServices регистрирует gRPC сервисы Inventory.
func RegisterServices(server *grpc.Server) {
	inventoryv1.RegisterInventoryServiceServer(server, NewInventoryAPI())
}
