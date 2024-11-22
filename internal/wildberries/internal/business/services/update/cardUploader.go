package update

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type CardUploader interface {
	Upload(data interface{}) (int, error)
}

const uploadUrl = "https://content-api.wildberries.ru/content/v2/cards/upload"

type CardUploaderImpl struct {
	log.Logger
	services.AuthEngine
}

func NewCardUploaderImpl(auth services.AuthEngine) *CardUploaderImpl {
	return &CardUploaderImpl{
		AuthEngine: auth,
	}
}

func (c *CardUploaderImpl) Upload(data []byte) (int, error) {
	body, length, err := c.uploadModels(uploadUrl, data)
	if err != nil {
		return 0, err
	}
	log.Printf("Uploaded : %s", body)
	return length, nil
}

func (c *CardUploaderImpl) PreloadCheck(data []byte) (interface{}, error) {
	var singleRequest request.CreateCardRequestData
	var requestArray []request.CreateCardRequestData

	// попробуем сначала разобрать как одиночный объект
	err := json.Unmarshal(data, &singleRequest)
	if err == nil {
		// если удалось, валидируем одиночный объект
		if validationErr := singleRequest.Validate(); validationErr != nil {
			return nil, validationErr
		}
		return singleRequest, nil
	}

	// если это не одиночный объект, пробуем массив
	err = json.Unmarshal(data, &requestArray)
	if err != nil {
		// Если оба варианта не подходят, возвращаем ошибку
		return nil, fmt.Errorf("invalid data format: %w", err)
	}

	// валидируем массив объектов
	for i, req := range requestArray {
		if validationErr := req.Validate(); validationErr != nil {
			return nil, fmt.Errorf("validation error in item %d: %w", i, validationErr)
		}
	}

	return requestArray, nil
}

func (c *CardUploaderImpl) uploadModels(url string, models interface{}) ([]byte, int, error) {
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
	c.SetApiKey(req)

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
