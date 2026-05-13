package converter

import (
	"github.com/yerkesh/inventory/internal/model"
	"github.com/yerkesh/inventory/internal/repository/record"
)

// PartToModel преобразует запись хранилища в доменную модель.
func PartToModel(part record.Part) model.Part {
	return model.Part{
		UUID:          part.UUID,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		PartType:      model.PartType(part.PartType),
		StockQuantity: part.StockQuantity,
		CreatedAt:     part.CreatedAt,
	}
}

// PartToRecord преобразует доменную модель в запись хранилища.
func PartToRecord(part model.Part) record.Part {
	return record.Part{
		UUID:          part.UUID,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		PartType:      string(part.PartType),
		StockQuantity: part.StockQuantity,
		CreatedAt:     part.CreatedAt,
	}
}
