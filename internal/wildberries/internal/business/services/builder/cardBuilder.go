package builder

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
	"gomarketplace_api/internal/wildberries/internal/business/models/get"
	"gomarketplace_api/pkg/business/service"
)

type CardBuilder struct {
	card        *get.WildberriesCard
	textService service.ITextService
}

func NewCardBuilder(textService service.ITextService) *CardBuilder {
	return &CardBuilder{
		card:        &get.WildberriesCard{},
		textService: textService,
	}
}

// FromNomenclature инициализирует базовые поля карточки из номенклатуры
func (b *CardBuilder) FromNomenclature(n response.Nomenclature) *CardBuilder {
	b.card.NmID = n.NmID
	b.card.VendorCode = n.VendorCode
	b.card.Brand = n.Brand
	b.card.Title = n.Title
	b.card.Characteristics = n.Characteristics
	b.card.Sizes = n.Sizes
	b.card.Dimensions = *n.Dimensions.Unwrap()
	return b
}

// WithUpdatedTitle обновляет название с учетом бренда
func (b *CardBuilder) WithUpdatedTitle(appellation string, maxLength int) *CardBuilder {
	title := b.textService.ClearAndReduce(appellation, maxLength)

	changedBrand := b.textService.ReplaceEngLettersToRus(b.card.Brand)
	if ok, newTitle := b.textService.ReplaceSymbols(title, map[string]string{b.card.Brand: changedBrand}); ok {
		title = newTitle
	} else if ok, newTitle := b.textService.FitIfPossible(title, changedBrand, maxLength); ok {
		title = newTitle
	}

	b.card.Title = title
	return b
}

// WithDescription обновляет описание
func (b *CardBuilder) WithDescription(description string, maxLength int) *CardBuilder {
	b.card.Description = b.textService.ClearAndReduce(description, maxLength)
	return b
}

// WithFallbackDescription использует appellation как описание, если основное описание отсутствует
func (b *CardBuilder) WithFallbackDescription(appellation string, maxLength int) *CardBuilder {
	b.card.Description = b.textService.ClearAndReduce(appellation, maxLength)
	return b
}

// Build возвращает готовую карточку
func (b *CardBuilder) Build() *get.WildberriesCard {
	return b.card
}

func (b *CardBuilder) Clear() {
	b.card = &get.WildberriesCard{}
}
