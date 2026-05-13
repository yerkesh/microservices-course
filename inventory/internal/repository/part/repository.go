package part

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	errs "github.com/yerkesh/inventory/internal/errors"
	"github.com/yerkesh/inventory/internal/model"
	"github.com/yerkesh/inventory/internal/repository/converter"
	"github.com/yerkesh/inventory/internal/repository/record"
)

// Repository хранит детали в памяти.
type Repository struct {
	mu    sync.RWMutex
	parts map[uuid.UUID]record.Part
}

// New создаёт репозиторий с seed-данными.
func New() *Repository {
	now := time.Now()

	return &Repository{
		parts: map[uuid.UUID]record.Part{
			uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"): {
				UUID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
				Name:          "Алюминиевый корпус",
				Description:   "Лёгкий корпус для небольших кораблей",
				Price:         500000,
				PartType:      string(model.PartTypeHull),
				StockQuantity: 10,
				CreatedAt:     now,
			},
			uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"): {
				UUID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
				Name:          "Титановый корпус",
				Description:   "Прочный корпус для средних кораблей",
				Price:         1500000,
				PartType:      string(model.PartTypeHull),
				StockQuantity: 5,
				CreatedAt:     now,
			},
			uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"): {
				UUID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
				Name:          "Ионный двигатель C",
				Description:   "Базовый ионный двигатель класса C",
				Price:         300000,
				PartType:      string(model.PartTypeEngine),
				StockQuantity: 8,
				CreatedAt:     now,
			},
			uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"): {
				UUID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
				Name:          "Ионный двигатель B",
				Description:   "Улучшенный ионный двигатель класса B",
				Price:         800000,
				PartType:      string(model.PartTypeEngine),
				StockQuantity: 3,
				CreatedAt:     now,
			},
			uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"): {
				UUID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"),
				Name:          "Энергетический щит",
				Description:   "Стандартный энергетический щит",
				Price:         400000,
				PartType:      string(model.PartTypeShield),
				StockQuantity: 6,
				CreatedAt:     now,
			},
			uuid.MustParse("550e8400-e29b-41d4-a716-446655440006"): {
				UUID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440006"),
				Name:          "Лазерная пушка",
				Description:   "Точная лазерная пушка",
				Price:         250000,
				PartType:      string(model.PartTypeWeapon),
				StockQuantity: 7,
				CreatedAt:     now,
			},
			uuid.MustParse("550e8400-e29b-41d4-a716-446655440007"): {
				UUID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440007"),
				Name:          "Карбоновый корпус",
				Description:   "Экспериментальный корпус, временно отсутствует на складе",
				Price:         2000000,
				PartType:      string(model.PartTypeHull),
				StockQuantity: 0,
				CreatedAt:     now,
			},
		},
	}
}

// Get возвращает деталь по UUID.
func (r *Repository) Get(_ context.Context, partUUID uuid.UUID) (model.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	part, ok := r.parts[partUUID]
	if !ok {
		return model.Part{}, errs.ErrPartNotFound
	}

	return converter.PartToModel(part), nil
}

// List возвращает детали по UUID или по типу.
func (r *Repository) List(_ context.Context, partUUIDs []uuid.UUID, partType model.PartType) ([]model.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(partUUIDs) > 0 {
		parts := make([]model.Part, 0, len(partUUIDs))
		for _, partUUID := range partUUIDs {
			part, ok := r.parts[partUUID]
			if !ok {
				return nil, errs.ErrPartNotFound
			}

			parts = append(parts, converter.PartToModel(part))
		}

		return parts, nil
	}

	parts := make([]model.Part, 0, len(r.parts))
	for _, part := range r.parts {
		if partType == model.PartTypeUnspecified || model.PartType(part.PartType) == partType {
			parts = append(parts, converter.PartToModel(part))
		}
	}

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].Name < parts[j].Name
	})

	return parts, nil
}
