package v1

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yerkesh/order/internal/client/grpc/inventory/v1/converter"
	errs "github.com/yerkesh/order/internal/errors"
	"github.com/yerkesh/order/internal/model"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
)

// Client оборачивает gRPC клиент InventoryService.
type Client struct {
	inventoryClient inventoryv1.InventoryServiceClient
}

// New создаёт клиент InventoryService.
func New(inventoryClient inventoryv1.InventoryServiceClient) *Client {
	return &Client{inventoryClient: inventoryClient}
}

// ListParts возвращает детали из InventoryService.
func (c *Client) ListParts(ctx context.Context, uuids []string) ([]model.Part, error) {
	resp, err := c.inventoryClient.ListParts(ctx, &inventoryv1.ListPartsRequest{
		Uuids: uuids,
	})
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return nil, errs.ErrPartNotFound
		case codes.InvalidArgument:
			return nil, errs.ErrInvalidUUID
		default:
			return nil, fmt.Errorf("получить список деталей: %w", err)
		}
	}

	return converter.PartsToModel(resp.GetParts()), nil
}
