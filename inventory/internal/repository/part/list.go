package part

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"

	errs "github.com/yerkesh/inventory/internal/errors"
	"github.com/yerkesh/inventory/internal/model"
	"github.com/yerkesh/inventory/internal/repository/converter"
	"github.com/yerkesh/inventory/internal/repository/record"
)

func (r *repository) List(ctx context.Context, uuids []uuid.UUID, partType model.PartType) ([]model.Part, error) {
	query := `
		SELECT uuid, name, description, price, part_type, stock_quantity, created_at
		FROM parts
	`
	var args []any
	if len(uuids) > 0 {
		query += " WHERE uuid = ANY($1)"
		args = append(args, uuids)
	} else if partType != model.PartTypeUnspecified {
		query += " WHERE part_type = $1"
		args = append(args, string(partType))
	}

	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("получить список деталей: %w", err)
	}
	defer rows.Close()

	var parts []model.Part
	for rows.Next() {
		var part record.Part
		if err := rows.Scan(
			&part.UUID,
			&part.Name,
			&part.Description,
			&part.Price,
			&part.PartType,
			&part.StockQuantity,
			&part.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("прочитать деталь: %w", err)
		}
		parts = append(parts, converter.PartToModel(part))
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать список деталей: %w", err)
	}

	if len(uuids) > 0 {
		partsByUUID := make(map[uuid.UUID]model.Part, len(parts))
		for _, part := range parts {
			partsByUUID[part.UUID] = part
		}

		orderedParts := make([]model.Part, 0, len(uuids))
		for _, partUUID := range uuids {
			part, ok := partsByUUID[partUUID]
			if !ok {
				return nil, errs.ErrPartNotFound
			}
			orderedParts = append(orderedParts, part)
		}

		return orderedParts, nil
	}

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].Name < parts[j].Name
	})

	return parts, nil
}
