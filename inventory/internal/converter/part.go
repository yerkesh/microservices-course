package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yerkesh/inventory/internal/model"
	inventoryv1 "github.com/yerkesh/shared/pkg/proto/inventory/v1"
)

// PartToProto преобразует доменную деталь в proto.
func PartToProto(part model.Part) *inventoryv1.Part {
	return &inventoryv1.Part{
		Uuid:          part.UUID.String(),
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		PartType:      PartTypeToProto(part.PartType),
		StockQuantity: part.StockQuantity,
		CreatedAt:     timestamppb.New(part.CreatedAt),
	}
}

// PartsToProto преобразует список доменных деталей в proto.
func PartsToProto(parts []model.Part) []*inventoryv1.Part {
	result := make([]*inventoryv1.Part, 0, len(parts))
	for _, part := range parts {
		result = append(result, PartToProto(part))
	}

	return result
}

// PartTypeToProto преобразует доменный тип детали в proto enum.
func PartTypeToProto(partType model.PartType) inventoryv1.PartType {
	switch partType {
	case model.PartTypeHull:
		return inventoryv1.PartType_PART_TYPE_HULL
	case model.PartTypeEngine:
		return inventoryv1.PartType_PART_TYPE_ENGINE
	case model.PartTypeShield:
		return inventoryv1.PartType_PART_TYPE_SHIELD
	case model.PartTypeWeapon:
		return inventoryv1.PartType_PART_TYPE_WEAPON
	default:
		return inventoryv1.PartType_PART_TYPE_UNSPECIFIED
	}
}

// PartTypeToModel преобразует proto enum в доменный тип детали.
func PartTypeToModel(partType inventoryv1.PartType) model.PartType {
	switch partType {
	case inventoryv1.PartType_PART_TYPE_HULL:
		return model.PartTypeHull
	case inventoryv1.PartType_PART_TYPE_ENGINE:
		return model.PartTypeEngine
	case inventoryv1.PartType_PART_TYPE_SHIELD:
		return model.PartTypeShield
	case inventoryv1.PartType_PART_TYPE_WEAPON:
		return model.PartTypeWeapon
	default:
		return model.PartTypeUnspecified
	}
}
