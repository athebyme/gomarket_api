package config

import (
	"flag"
	"log"
	"os"
)

type Config interface{}

type MarketplaceConfig interface {
	ApiKey() string
}

type WildberriesConfig struct {
	apiKey string
}

func (cfg *WildberriesConfig) ApiKey() string {
	return cfg.apiKey
}

func GetWildberriesConfig() *WildberriesConfig {
	apiKey := flag.String("WB_API_KEY", "", "Wildberries API key")
	if *apiKey == "" {
		if envApiKey, exists := os.LookupEnv("WB_API_KEY"); exists {
			apiKey = &envApiKey
		} else {
			log.Fatal("wildberries api key is required")
		}
	}
	return &WildberriesConfig{apiKey: *apiKey}
}
