package v1

import (
	"context"

	"github.com/yerkesh/inventory/internal/converter"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
)

// ListParts возвращает список деталей.
func (a *API) ListParts(ctx context.Context, req *inventoryv1.ListPartsRequest) (*inventoryv1.ListPartsResponse, error) {
	parts, err := a.partService.List(ctx, req.GetUuids(), converter.PartTypeToModel(req.GetPartType()))
	if err != nil {
		return nil, partErrorToStatus(err)
	}

	return &inventoryv1.ListPartsResponse{Parts: converter.PartsToProto(parts)}, nil
}
