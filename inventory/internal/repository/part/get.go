package part

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	errs "github.com/yerkesh/inventory/internal/errors"
	"github.com/yerkesh/inventory/internal/model"
	"github.com/yerkesh/inventory/internal/repository/converter"
	"github.com/yerkesh/inventory/internal/repository/record"
)

func (r *repository) Get(ctx context.Context, uuid uuid.UUID) (model.Part, error) {
	var part record.Part

	const query = `
		SELECT uuid, name, description, price, part_type, stock_quantity, created_at
		FROM parts
		WHERE uuid = $1
	`
	err := r.getter.DefaultTrOrDB(ctx, r.pool).QueryRow(ctx, query, uuid).Scan(
		&part.UUID,
		&part.Name,
		&part.Description,
		&part.Price,
		&part.PartType,
		&part.StockQuantity,
		&part.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Part{}, errs.ErrPartNotFound
		}

		return model.Part{}, fmt.Errorf("получить деталь: %w", err)
	}

	return converter.PartToModel(part), nil
}
