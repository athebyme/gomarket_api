package handlers

import (
	"context"
	"gomarketplace_api/internal/wildberries/business/services/update"
)

type UpdateHandler interface {
	Handle(ctx context.Context, processor *update.CardProcessor) error
}

type TitleUpdateHandler struct {
	next UpdateHandler
}

func (h *TitleUpdateHandler) Handle(ctx context.Context, processor *update.CardProcessor) error {
	if processor.appellation != "" {
		processor.nomenclature.Title = processor.appellation
	}
	if h.next != nil {
		return h.next.Handle(ctx, processor)
	}
	return nil
}

type DescriptionUpdateHandler struct {
	next UpdateHandler
}

func (h *DescriptionUpdateHandler) Handle(ctx context.Context, processor *update.CardProcessor) error {
	if processor.description != "" {
		processor.nomenclature.Description = processor.description
	}
	if h.next != nil {
		return h.next.Handle(ctx, processor)
	}
	return nil
}
