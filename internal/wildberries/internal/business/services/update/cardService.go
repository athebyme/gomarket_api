package update

import (
	"fmt"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wildberries/internal/business/services/builder"
	"gomarketplace_api/internal/wildberries/internal/business/services/parse"
	clients2 "gomarketplace_api/internal/wildberries/pkg/clients"
	"gomarketplace_api/pkg/business/service"
	"gomarketplace_api/pkg/logger"
	"io"
	"sync"
)

type CardService struct {
	cardBuilder  builder.Proxy
	textService  service.ITextService
	brandService parse.BrandService
	wsclient     *clients2.WServiceClient

	config.WildberriesConfig
	logger.Logger
}

func NewCardService(wsClientUrl string, textService service.ITextService, writer io.Writer, wildberriesConfig config.WildberriesConfig) *CardService {
	_log := logger.NewLogger(writer, "[CardService]")
	cardBuilder := parse.NewCardBuilderEngine(writer)

	return &CardService{
		Logger:            _log,
		cardBuilder:       cardBuilder,
		WildberriesConfig: wildberriesConfig,
		brandService:      parse.NewBrandServiceWildberries(wildberriesConfig.WbBanned.BannedBrands),
		wsclient:          clients2.NewWServiceClient(wsClientUrl, writer),
		textService:       textService,
	}
}

func (s *CardService) Prepare(ids []int) (interface{}, error) {
	s.Logger.SetPrefix(" - Preparation stage - ")

	var filtered sync.Map

	globalIDsMap, err := s.filterGlobalIDs(ids)
	if err != nil {
		return nil, err
	}
	s.logFilteredIDs(&filtered, ids, globalIDsMap, "Global ID filtering")
	ids = s.extractKeys(globalIDsMap)

	brandsMap, err := s.filterBrands(ids)
	if err != nil {
		return nil, err
	}
	s.logFilteredIDs(&filtered, ids, brandsMap, "Brand filtering")

	ids = s.extractKeys(brandsMap)

	appellationsMap, err := s.filterAppellations(ids)
	if err != nil {
		return nil, err
	}
	s.logFilteredIDs(&filtered, ids, appellationsMap, "Appellation filtering")

	ids = s.extractKeys(appellationsMap)

	descriptionsMap, err := s.filterDescriptions(ids)
	if err != nil {
		return nil, err
	}
	s.logFilteredIDs(&filtered, ids, descriptionsMap, "Description filtering")

	ids = s.extractKeys(appellationsMap)

	filtered.Range(func(key, value any) bool {
		s.Log("Excluded ID: %v, Reason: %v\n", key, value)
		return true
	})

	s.Logger.SetPrefix("")
	return ids, nil
}

func (s *CardService) filterAppellations(ids []int) (map[int]interface{}, error) {
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

	filtered := make(map[int]interface{}, len(ids))
	for id, app := range appellations {
		if _, exists := idSet[id]; exists {
			switch app.(type) {
			case string:
				filtered[id] = s.textService.ClearAndReduce(app.(string), 60)
			default:
				return nil, fmt.Errorf("unsupported type of appellation ")
			}
		}
	}

	return filtered, nil
}

func (s *CardService) filterDescriptions(ids []int) (map[int]interface{}, error) {
	idSet := make(map[int]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	descriptions, err := s.wsclient.FetchDescriptions()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return descriptions, nil
	}

	filtered := make(map[int]interface{}, len(ids))
	for id, desc := range descriptions {
		if _, exists := idSet[id]; exists {
			switch desc.(type) {
			case string:
				filtered[id] = s.textService.ClearAndReduce(desc.(string), 2000)
			default:
				return nil, fmt.Errorf("unsupported type of description ")
			}
		}
	}

	return filtered, nil
}

func (s *CardService) filterGlobalIDs(ids []int) (map[int]interface{}, error) {
	idSet := make(map[int]interface{})
	if len(ids) == 0 {
		for _, id := range ids {
			idSet[id] = struct{}{}
		}
	} else {
		globalIDs, err := s.wsclient.FetchGlobalIDs()
		if err != nil {
			return nil, err
		}
		for _, id := range globalIDs {
			idSet[id] = struct{}{}
		}
	}
	return idSet, nil
}

func (s *CardService) filterBrands(ids []int) (map[int]interface{}, error) {
	idSet := make(map[int]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	brands, err := s.wsclient.FetchBrands()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return brands, nil
	}

	filtered := make(map[int]interface{}, len(ids))
	for id, brand := range brands {
		if _, exists := idSet[id]; exists {
			switch brand.(type) {
			case string:
				strBrand := brand.(string)
				if s.brandService.IsBanned(strBrand) {
					continue
				}
				filtered[id] = brand.(string)
			default:
				return nil, fmt.Errorf("unsupported type of brand ")
			}
		}
	}

	return filtered, nil
}

func (s *CardService) filterPrice(ids []int) (map[int]interface{}, error) {
	idSet := make(map[int]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	prices, err := s.wsclient.FetchPrices()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return prices, nil
	}

	filtered := make(map[int]interface{}, len(ids))
	for id, brand := range prices {
		if _, exists := idSet[id]; exists {
			switch brand.(type) {
			case string:
				strBrand := brand.(string)
				if s.brandService.IsBanned(strBrand) {
					continue
				}
				filtered[id] = brand.(string)
			default:
				return nil, fmt.Errorf("unsupported type of brand ")
			}
		}
	}

	return filtered, nil
}

// logFilteredIDs записывает исключённые артикулы в sync.Map
func (s *CardService) logFilteredIDs(filtered *sync.Map, originalIDs []int, filteredMap map[int]interface{}, reason string) {
	originalSet := make(map[int]struct{})
	for _, id := range originalIDs {
		originalSet[id] = struct{}{}
	}

	// Ищем отсеянные элементы
	for id := range originalSet {
		if _, exists := filteredMap[id]; !exists {
			filtered.Store(id, reason) // Записываем причину отсеивания
		}
	}
}

// extractKeys извлекает ключи из мапы
func (s *CardService) extractKeys(data map[int]interface{}) []int {
	keys := make([]int, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}
