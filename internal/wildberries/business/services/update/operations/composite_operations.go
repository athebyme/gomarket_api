package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/business/models/dto/response"
	"gomarketplace_api/internal/wildberries/business/services/update/operations/domain/models"
)

// CompositeOperation позволяет объединить несколько операций обновления
type CompositeOperation struct {
	operations []UpdateOperation
	// Опциональные настройки для композитного обновления
	settings CompositeSettings
}

type CompositeSettings struct {
	// Продолжать ли обработку если одна из операций завершилась с ошибкой
	ContinueOnError bool
	// Объединять ли результаты всех операций в один запрос
	MergeResults bool
}

// NewCompositeOperation создает новую композитную операцию
func NewCompositeOperation(operations []UpdateOperation, settings CompositeSettings) *CompositeOperation {
	return &CompositeOperation{
		operations: operations,
		settings:   settings,
	}
}

// Validate проверяет номенклатуру для всех операций
func (c *CompositeOperation) Validate(nom response.Nomenclature) bool {
	for _, op := range c.operations {
		if !op.Validate(nom) {
			return false
		}
	}
	return true
}

// Process обрабатывает номенклатуру всеми операциями
func (c *CompositeOperation) Process(ctx context.Context, nom response.Nomenclature) (request.Model, error) {
	if c.settings.MergeResults {
		return c.processMerged(ctx, nom)
	}
	return c.processSequential(ctx, nom)
}

// processMerged объединяет результаты всех операций в один запрос
func (c *CompositeOperation) processMerged(ctx context.Context, nom response.Nomenclature) (request.Model, error) {
	var compositeModel models.CompositeModel

	for _, op := range c.operations {
		model, err := op.Process(ctx, nom)
		if err != nil {
			if !c.settings.ContinueOnError {
				return nil, fmt.Errorf("operation failed: %w", err)
			}
			continue
		}

		if err := compositeModel.Merge(model); err != nil {
			if !c.settings.ContinueOnError {
				return nil, fmt.Errorf("merge failed: %w", err)
			}
		}
	}

	return compositeModel, nil
}

// processSequential обрабатывает операции последовательно
func (c *CompositeOperation) processSequential(ctx context.Context, nom response.Nomenclature) (request.Model, error) {
	var models []request.Model

	for _, op := range c.operations {
		model, err := op.Process(ctx, nom)
		if err != nil {
			if !c.settings.ContinueOnError {
				return nil, fmt.Errorf("operation failed: %w", err)
			}
			continue
		}
		models = append(models, model)
	}

	return &SequentialModel{Models: models}, nil
}

// SequentialModel хранит модели для последовательной обработки
type SequentialModel struct {
	Models []request.Model
}

// SequentialModel реализует интерфейс Model
func (s SequentialModel) ToBytes() ([]byte, error) {
	type sequentialRequest struct {
		Requests []json.RawMessage `json:"requests"`
	}

	req := sequentialRequest{
		Requests: make([]json.RawMessage, 0, len(s.Models)),
	}

	for _, model := range s.Models {
		bytes, err := model.ToBytes()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize model: %w", err)
		}
		req.Requests = append(req.Requests, bytes)
	}

	return json.Marshal(req)
}
