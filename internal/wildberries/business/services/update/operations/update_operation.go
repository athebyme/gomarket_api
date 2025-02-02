package operations

import (
	"context"
	"gomarketplace_api/internal/wildberries/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/business/models/dto/response"
)

type UpdateOperation interface {
	// Validate проверяет, подходит ли данная номенклатура для операции.
	Validate(nom response.Nomenclature) bool
	// Process обрабатывает номенклатуру и возвращает модель запроса для обновления.
	Process(ctx context.Context, nom response.Nomenclature) (request.Model, error)
}
