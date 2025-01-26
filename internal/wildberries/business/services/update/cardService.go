package update

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wholesaler/pkg/requests"
	"gomarketplace_api/internal/wildberries/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/business/services"
	"gomarketplace_api/internal/wildberries/business/services/builder"
	parse2 "gomarketplace_api/internal/wildberries/business/services/parse"
	clients2 "gomarketplace_api/internal/wildberries/pkg/clients"
	"gomarketplace_api/pkg/business/service"
	"gomarketplace_api/pkg/logger"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const uploadCardsUrl = "https://content-api.wildberries.ru/content/v2/cards/upload"

type CardService struct {
	cardBuilder  builder.Proxy
	textService  service.ITextService
	brandService parse2.BrandService
	nmService    NomenclatureService
	wsclient     *clients2.WServiceClient

	config.WildberriesConfig
	logger.Logger
	services.AuthEngine
}

func NewCardService(
	wsClientUrl string,
	textService service.ITextService,
	writer io.Writer,
	wildberriesConfig config.WildberriesConfig) *CardService {
	_log := logger.NewLogger(writer, "[CardService]")
	cardBuilder := parse2.NewCardBuilderEngine(writer, wildberriesConfig.WbValues)

	return &CardService{
		Logger:            _log,
		cardBuilder:       cardBuilder,
		WildberriesConfig: wildberriesConfig,
		brandService:      parse2.NewBrandServiceWildberries(wildberriesConfig.WbBanned.BannedBrands),
		wsclient:          clients2.NewWServiceClient(wsClientUrl, writer),
		textService:       textService,
		AuthEngine:        services.NewBearerAuth(wildberriesConfig.ApiKey),
	}
}

func (s *CardService) Prepare(ids []int) (interface{}, error) {
	preparationLogger := s.Logger.WithPrefix("[Preparation stage] ")
	startTime := time.Now()

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

	ids = s.extractKeys(descriptionsMap)

	barcodesMap, err := s.filterBarcodes(ids)
	if err != nil {
		return nil, err
	}
	s.logFilteredIDs(&filtered, ids, barcodesMap, "Barcodes filtering")

	ids = s.extractKeys(barcodesMap)

	filtered.Range(func(key, value any) bool {
		preparationLogger.Log("Excluded ID: %v, Reason: %v", key, value)
		return true
	})

	totalTime := time.Since(startTime)
	preparationLogger.Log("Time spent on preparations : %s", totalTime.String())

	return ids, nil
}

func (s *CardService) PrepareAndUpload(ids []int) (interface{}, error) {
	preparingResult, err := s.Prepare(ids)
	if err != nil {
		return nil, err
	}

	ids, ok := preparingResult.([]int)
	if !ok {
		return nil, fmt.Errorf("PrepareAndUpload returned invalid result")
	}

	appellations, err := s.filterAppellations(ids)
	if err != nil {
		return nil, err
	}

	descriptions, err := s.filterDescriptions(ids)
	if err != nil {
		return nil, err
	}

	brands, err := s.filterBrands(ids)
	if err != nil {
		return nil, err
	}

	prices, err := s.filterPrice(ids)
	if err != nil {
		return nil, err
	}

	var cards []request.CreateCardRequestData
	for _, id := range ids {
		card, err := s.cardBuilder.WithBrand(brands[id].(string)).
			WithDescription(descriptions[id].(string)).
			WithTitle(appellations[id].(string)).
			WithVendorCode(fmt.Sprintf("id-%d-%d", id, s.WbIdentity.Code)).
			WithPrice(prices[id].(int) * 2).
			Build()
		if err != nil {
			return nil, err
		}
		cards = append(cards, *card.(*request.CreateCardRequestData))
	}
	return cards, nil
}

func (s *CardService) SendToServerModels(models interface{}) ([]byte, int, error) {
	return s.sendToServer(uploadCardsUrl, models)
}

func (cu *CardService) sendToServer(url string, models interface{}) ([]byte, int, error) {
	log.Printf("Sending models to server...")

	requestBody, err := json.Marshal(models)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to marshal update request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, 500, fmt.Errorf("failed to create update request: %w", err)
	}

	requestBodySize := len(requestBody)
	requestBodySizeMB := float64(requestBodySize) / (1 << 20)
	log.Printf("Request Body Size: %d bytes (%.2f MB)\n", requestBodySize, requestBodySizeMB)

	req.Header.Set("Content-Type", "application/json")
	cu.SetApiKey(req)

	log.Printf("Sending request body: %v", string(requestBody))
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to upload models: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &errorResponse); err != nil {
			return nil, resp.StatusCode, fmt.Errorf("failed to unmarshal error response: %w", err)
		}

		errorDetails, err := json.MarshalIndent(errorResponse, "", "  ")
		if err != nil {
			return nil, resp.StatusCode, fmt.Errorf("failed to format error details: %w", err)
		}

		log.Printf("Upload failed with status: %d, error details: %s", resp.StatusCode, string(errorDetails))
		return bodyBytes, resp.StatusCode, fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	return bodyBytes, resp.StatusCode, nil
}

func (s *CardService) filterAppellations(ids []int) (map[int]interface{}, error) {
	idSet := make(map[int]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	appellations, err := s.wsclient.FetchAppellations(requests.AppellationsRequest{FilterRequest: requests.FilterRequest{
		ProductIDs: ids,
	}})

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

	descriptions, err := s.wsclient.FetchDescriptions(requests.DescriptionRequest{FilterRequest: requests.FilterRequest{ProductIDs: ids}, IncludeEmptyDescriptions: false})
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
	globalIDs, err := s.wsclient.FetchGlobalIDs()
	if err != nil {
		return nil, err
	}

	globalIDsMap := make(map[int]interface{}, len(ids))
	for _, id := range globalIDs {
		globalIDsMap[id] = struct{}{}
	}

	idSet := make(map[int]interface{})
	if len(ids) == 0 {
		return globalIDsMap, nil
	} else {
		for _, id := range ids {
			if _, ok := globalIDsMap[id]; ok {
				idSet[id] = struct{}{}
			}
		}
	}
	return idSet, nil
}

func (s *CardService) filterBrands(ids []int) (map[int]interface{}, error) {
	idSet := make(map[int]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	brands, err := s.wsclient.FetchBrands(requests.BrandRequest{FilterRequest: requests.FilterRequest{
		ProductIDs: ids,
	}})
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
				if s.brandService.IsBanned(strBrand) || strBrand == "" {
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

	prices, err := s.wsclient.FetchPrices(requests.PriceRequest{FilterRequest: requests.FilterRequest{
		ProductIDs: ids,
	}})
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return prices, nil
	}

	filtered := make(map[int]interface{}, len(ids))
	for id, price := range prices {
		if _, exists := idSet[id]; exists {
			switch price.(type) {
			case map[string]interface{}:
				priceResult := price.(map[string]interface{}) // Приведение к map[string]interface{}
				zValue, ok := priceResult["Z"].(float64)      // Пробуем получить значение "Z" как float64
				if !ok {
					return nil, fmt.Errorf("key 'Z' is missing or not a float64")
				}
				filtered[id] = int(zValue * 1.15)
			case float64, float32:
				filtered[id] = int(price.(float64))
			default:
				return nil, fmt.Errorf("unsupported type of price ")
			}
		}
	}

	return filtered, nil
}

func (s *CardService) filterBarcodes(ids []int) (map[int]interface{}, error) {
	barcodesRepo, err := s.wsclient.FetchBarcodes(requests.BarcodeRequest{FilterRequest: requests.FilterRequest{ProductIDs: ids}})
	if err != nil {
		return nil, err
	}

	barcodes := make(map[int]interface{}, len(ids))
	for id, barcode := range barcodesRepo {
		barcodes[id] = barcode
	}

	idSet := make(map[int]interface{})
	if len(ids) == 0 {
		return barcodes, nil
	} else {
		for _, id := range ids {
			switch v := barcodes[id].(type) {
			case string:
				idSet[id] = []string{v}
			case []string:
				idSet[id] = v
			case []interface{}:
				strSlice := make([]string, len(v))
				for i, item := range v {
					if str, ok := item.(string); ok {
						strSlice[i] = str
					}
				}
				idSet[id] = strSlice
			default:
				return nil, fmt.Errorf("unsupported type of barcode ")
			}
		}
	}
	return idSet, nil
}

func (s *CardService) filterSizes(ids []int) (map[int]interface{}, error) {
	sizes, err := s.wsclient.FetchSizes(requests.SizeRequest{FilterRequest: requests.FilterRequest{
		ProductIDs: ids,
	}})
	if err != nil {
		return nil, err
	}

	sizesMap := make(map[int]interface{}, len(ids))
	for id, _ := range sizes {
		sizesMap[id] = struct{}{}
	}

	idSet := make(map[int]interface{})
	if len(ids) == 0 {
		return sizesMap, nil
	} else {
		for _, id := range ids {
			if _, ok := sizesMap[id]; ok {
				idSet[id] = struct{}{}
			}
		}
	}
	return idSet, nil
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
