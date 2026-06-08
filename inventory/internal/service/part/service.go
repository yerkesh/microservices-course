package part

import (
	"context"

	"github.com/google/uuid"

	errs "github.com/yerkesh/inventory/internal/errors"
	"github.com/yerkesh/inventory/internal/model"
)

// Service реализует бизнес-логику каталога деталей.
type Service struct {
	txManager TxManager
	repo      Repository
}

// New создаёт сервис деталей.
func New(txManager TxManager, repo Repository) *Service {
	return &Service{txManager: txManager, repo: repo}
}

// Get возвращает деталь по UUID.
func (s *Service) Get(ctx context.Context, rawUUID string) (model.Part, error) {
	partUUID, err := parseUUID(rawUUID)
	if err != nil {
		return model.Part{}, err
	}

	var part model.Part
	if err = s.txManager.Do(ctx, func(ctx context.Context) error {
		var err error
		part, err = s.repo.Get(ctx, partUUID)
		return err
	}); err != nil {
		return model.Part{}, err
	}

	return part, nil
}

// List возвращает список деталей.
func (s *Service) List(ctx context.Context, rawUUIDs []string, partType model.PartType) ([]model.Part, error) {
	partUUIDs := make([]uuid.UUID, 0, len(rawUUIDs))
	for _, rawUUID := range rawUUIDs {
		partUUID, err := parseUUID(rawUUID)
		if err != nil {
			return nil, err
		}

		partUUIDs = append(partUUIDs, partUUID)
	}

	var parts []model.Part
	if err := s.txManager.Do(ctx, func(ctx context.Context) error {
		var err error
		parts, err = s.repo.List(ctx, partUUIDs, partType)
		return err
	}); err != nil {
		return nil, err
	}

	return parts, nil
}

func parseUUID(rawUUID string) (uuid.UUID, error) {
	if rawUUID == "" {
		return uuid.Nil, errs.ErrInvalidUUID
	}

	parsed, err := uuid.Parse(rawUUID)
	if err != nil || parsed == uuid.Nil {
		return uuid.Nil, errs.ErrInvalidUUID
	}

	return parsed, nil
}
