package service

import (
	"github.com/yerkesh/inventory/pkg/app"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
)

// NewInventoryServer создаёт gRPC сервер InventoryService.
func NewInventoryServer() inventoryv1.InventoryServiceServer {
	return app.NewInventoryAPI()
}
