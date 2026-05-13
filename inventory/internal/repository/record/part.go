package record

import (
	"time"

	"github.com/google/uuid"
)

// Part хранит деталь в формате in-memory хранилища.
type Part struct {
	UUID          uuid.UUID
	Name          string
	Description   string
	Price         int64
	PartType      string
	StockQuantity int64
	CreatedAt     time.Time
}
