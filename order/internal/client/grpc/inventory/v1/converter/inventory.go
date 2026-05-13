package converter

import (
	"github.com/google/uuid"

	"github.com/yerkesh/order/internal/model"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
)

// PartsToModel преобразует proto-детали в доменную модель OrderService.
func PartsToModel(parts []*inventoryv1.Part) []model.Part {
	result := make([]model.Part, 0, len(parts))
	for _, part := range parts {
		parsedUUID, err := uuid.Parse(part.GetUuid())
		if err != nil {
			continue
		}

		result = append(result, model.Part{
			UUID:          parsedUUID,
			Price:         part.GetPrice(),
			StockQuantity: part.GetStockQuantity(),
		})
	}

	return result
}
