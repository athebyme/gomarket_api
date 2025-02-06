package domain

import (
	"context"
	"errors"
	"fmt"
	clients2 "gomarketplace_api/internal/suppliers/wholesaler/pkg/clients"
	"gomarketplace_api/internal/wildberries/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/business/models/dto/response"
	"gomarketplace_api/internal/wildberries/business/services/update/operations/domain/models"
	"gomarketplace_api/internal/wildberries/pkg/clients"
)

const (
	MediaUploadURL = "https://content-api.wildberries.ru/content/v3/media/save"
)

var (
	ErrMediaFilesContainsMoreData = errors.New("media files contains more than file than base has")
)

type MediaUpdateOperation struct {
	mediaMap map[int][]string
	client   *clients.WServiceClient
}

func NewMediaUpdateOperation(client *clients.WServiceClient) *MediaUpdateOperation {
	return &MediaUpdateOperation{
		mediaMap: make(map[int][]string),
		client:   client,
	}
}

// Validate проверяет, что:
// 1. Номенклатура имеет ровно 1 фотографию (для обновления).
// 2. Номенклатура имеет корректный globalID.
// 3. Для globalID есть ссылки в mediaMap.
func (op *MediaUpdateOperation) Validate(nom response.Nomenclature) bool {
	if len(nom.Photos) <= 1 {
		return false
	}
	globalID, err := nom.GlobalID()
	if err != nil || globalID == 0 {
		return false
	}
	urls, ok := op.mediaMap[globalID]
	if !ok || len(urls) == 0 {
		return false
	}
	return true
}

// Process создаёт модель запроса на основе номенклатуры и mediaMap.
// Если в mediaMap для globalID только 1 URL – дублирует его для улучшения качества.
func (op *MediaUpdateOperation) Process(ctx context.Context, nom response.Nomenclature) (request.Model, error) {
	globalID, err := nom.GlobalID()
	if err != nil {
		return nil, fmt.Errorf("invalid globalID: %w", err)
	}
	urls := op.mediaMap[globalID]
	if len(urls) < len(nom.Photos) {
		return nil, ErrMediaFilesContainsMoreData
	}
	if len(urls) == 1 {
		urls = append(urls, urls[0])
	}
	urls = append(urls, "http://media.athebyme-market.ru/anonymous/package/image/png")
	model := models.MediaModel{NmID: nom.NmID, URLs: urls}
	return model, nil
}

func (op *MediaUpdateOperation) MediaUrls(ctx context.Context, censored bool) (int, error) {
	mediaRequest := clients2.ImageRequest{
		Censored: censored,
	}
	mediaMapRaw, err := op.client.FetcherChain.Fetch(ctx, "media", mediaRequest)
	if err != nil {
		return 0, fmt.Errorf("error fetching media urls: %w", err)
	}

	mediaMap, ok := mediaMapRaw.(map[int][]string)
	if !ok {
		return 0, fmt.Errorf("unexpected type for mediaMap: expected map[int][]string, got %T", mediaMapRaw)
	}

	op.mediaMap = mediaMap
	return len(mediaMap), nil
}
