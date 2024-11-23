package update

import (
	"gomarketplace_api/internal/wildberries/internal/business/services/builder"
	"gomarketplace_api/internal/wildberries/internal/business/services/parse"
	clients2 "gomarketplace_api/internal/wildberries/pkg/clients"
	"gomarketplace_api/pkg/logger"
	"io"
)

type CardService struct {
	cardBuilder builder.Proxy
	wsclient    *clients2.WServiceClient
	logger.Logger
}

func NewCardService(wsClientUrl string, writer io.Writer) *CardService {
	_log := logger.NewLogger(writer, "[CardService]")
	cardBuilder := parse.NewCardBuilderEngine(writer)

	return &CardService{
		Logger:      _log,
		cardBuilder: cardBuilder,
		wsclient:    clients2.NewWServiceClient(wsClientUrl, writer),
	}
}

func (s *CardService) FetchCards(ids []int) (interface{}, error) {
	appellations, err := s.filterAppellations(ids)
	if err != nil {
		return nil, err
	}
}

func (s *CardService) filterAppellations(ids []int) (interface{}, error) {
	idSet := make(map[int]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	appellations, err := s.wsclient.FetchAppellations()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return appellations, nil
	}

	filteredAppellations := make([]string, 0, len(ids))
	for id, app := range appellations {
		if _, exists := idSet[id]; exists {
			filteredAppellations = append(filteredAppellations, app)
		}
	}

	return filteredAppellations, nil
}
