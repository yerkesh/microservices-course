package record

import (
	"time"

	"github.com/google/uuid"
)

// Part хранит деталь в формате PostgreSQL-хранилища.
type Part struct {
	UUID          uuid.UUID `db:"uuid"`
	Name          string    `db:"name"`
	Description   string    `db:"description"`
	Price         int64     `db:"price"`
	PartType      string    `db:"part_type"`
	StockQuantity int64     `db:"stock_quantity"`
	CreatedAt     time.Time `db:"created_at"`
}
