package v1

import (
	"context"

	"github.com/yerkesh/inventory/internal/model"
)

// PartService определяет контракт бизнес-логики деталей.
type PartService interface {
	Get(ctx context.Context, rawUUID string) (model.Part, error)
	List(ctx context.Context, rawUUIDs []string, partType model.PartType) ([]model.Part, error)
}
