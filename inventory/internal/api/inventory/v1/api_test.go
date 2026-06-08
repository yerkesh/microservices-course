package v1

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/yerkesh/inventory/internal/errors"
	"github.com/yerkesh/inventory/internal/model"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
)

type fakePartService struct {
	part  model.Part
	parts []model.Part
	err   error
}

func (s fakePartService) Get(_ context.Context, _ string) (model.Part, error) {
	return s.part, s.err
}

func (s fakePartService) List(_ context.Context, _ []string, _ model.PartType) ([]model.Part, error) {
	return s.parts, s.err
}

func TestGetPartSuccess(t *testing.T) {
	partUUID := uuid.New()
	api := New(fakePartService{part: model.Part{
		UUID:          partUUID,
		Name:          "Корпус",
		Price:         500000,
		PartType:      model.PartTypeHull,
		StockQuantity: 1,
	}})

	resp, err := api.GetPart(context.Background(), &inventoryv1.GetPartRequest{Uuid: partUUID.String()})

	require.NoError(t, err)
	require.Equal(t, partUUID.String(), resp.GetPart().GetUuid())
	require.Equal(t, inventoryv1.PartType_PART_TYPE_HULL, resp.GetPart().GetPartType())
}

func TestGetPartNotFound(t *testing.T) {
	api := New(fakePartService{err: errs.ErrPartNotFound})

	_, err := api.GetPart(context.Background(), &inventoryv1.GetPartRequest{Uuid: uuid.NewString()})

	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestListPartsInvalidUUID(t *testing.T) {
	api := New(fakePartService{err: errs.ErrInvalidUUID})

	_, err := api.ListParts(context.Background(), &inventoryv1.ListPartsRequest{Uuids: []string{"не-uuid"}})

	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
