package services

import (
	"context"
	"gomarketplace_api/internal/suppliers/wholesaler/pkg/requests"
	request2 "gomarketplace_api/internal/wildberries/business/models/dto/request"
	response2 "gomarketplace_api/internal/wildberries/business/models/dto/response"
)

type NomenclatureService interface {
	GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(
		ctx context.Context,
		settings request2.Settings,
		locale string,
		nomenclatureCh chan<- response2.Nomenclature,
	) error
}

type DataFetcher interface {
	FetchAppellations(ctx context.Context, request requests.AppellationsRequest) (map[int]interface{}, error)
	FetchDescriptions(ctx context.Context, request requests.AppellationsRequest) (map[int]interface{}, error)
}

type CardUpdater interface {
	UpdateCards(ctx context.Context, cards []request2.Model) (int, error)
}
