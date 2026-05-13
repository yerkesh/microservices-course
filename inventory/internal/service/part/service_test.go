package part

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	errs "github.com/yerkesh/inventory/internal/errors"
	"github.com/yerkesh/inventory/internal/model"
)

type fakeRepository struct {
	parts []model.Part
	err   error
}

func (r fakeRepository) Get(_ context.Context, partUUID uuid.UUID) (model.Part, error) {
	if r.err != nil {
		return model.Part{}, r.err
	}

	return model.Part{UUID: partUUID, Price: 500000}, nil
}

func (r fakeRepository) List(_ context.Context, partUUIDs []uuid.UUID, _ model.PartType) ([]model.Part, error) {
	if r.err != nil {
		return nil, r.err
	}

	if len(r.parts) > 0 {
		return r.parts, nil
	}

	return []model.Part{{UUID: partUUIDs[0], Price: 500000}}, nil
}

func TestServiceGetSuccess(t *testing.T) {
	partUUID := uuid.New()
	s := New(fakeRepository{})

	part, err := s.Get(context.Background(), partUUID.String())

	require.NoError(t, err)
	require.Equal(t, partUUID, part.UUID)
}

func TestServiceGetInvalidUUID(t *testing.T) {
	s := New(fakeRepository{})

	_, err := s.Get(context.Background(), "не-uuid")

	require.ErrorIs(t, err, errs.ErrInvalidUUID)
}

func TestServiceListInvalidUUID(t *testing.T) {
	s := New(fakeRepository{})

	_, err := s.List(context.Background(), []string{uuid.NewString(), "не-uuid"}, model.PartTypeUnspecified)

	require.ErrorIs(t, err, errs.ErrInvalidUUID)
}

func TestServiceListRepositoryError(t *testing.T) {
	repoErr := errors.New("репозиторий недоступен")
	s := New(fakeRepository{err: repoErr})

	_, err := s.List(context.Background(), []string{uuid.NewString()}, model.PartTypeUnspecified)

	require.ErrorIs(t, err, repoErr)
}
