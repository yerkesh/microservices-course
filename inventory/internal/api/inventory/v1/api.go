package v1

import inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"

// API реализует gRPC API InventoryService.
type API struct {
	inventoryv1.UnimplementedInventoryServiceServer
	partService PartService
}

// New создаёт API слой InventoryService.
func New(partService PartService) *API {
	return &API{partService: partService}
}
