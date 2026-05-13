package part

import (
	"context"

	"github.com/google/uuid"

	"github.com/yerkesh/inventory/internal/model"
)

// Repository определяет контракт хранилища деталей.
type Repository interface {
	Get(ctx context.Context, partUUID uuid.UUID) (model.Part, error)
	List(ctx context.Context, partUUIDs []uuid.UUID, partType model.PartType) ([]model.Part, error)
}
