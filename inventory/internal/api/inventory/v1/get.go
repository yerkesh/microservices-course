package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yerkesh/inventory/internal/converter"
	errs "github.com/yerkesh/inventory/internal/errors"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
)

// GetPart возвращает деталь по UUID.
func (a *API) GetPart(ctx context.Context, req *inventoryv1.GetPartRequest) (*inventoryv1.GetPartResponse, error) {
	part, err := a.partService.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, partErrorToStatus(err)
	}

	return &inventoryv1.GetPartResponse{Part: converter.PartToProto(part)}, nil
}

func partErrorToStatus(err error) error {
	switch {
	case errors.Is(err, errs.ErrInvalidUUID):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, errs.ErrPartNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Errorf(codes.Internal, "внутренняя ошибка: %v", err)
	}
}
