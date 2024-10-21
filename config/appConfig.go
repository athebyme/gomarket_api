package config

type MarketplaceConfig struct {
	ApiKey string
}

func GetMarketplaceConfig() *MarketplaceConfig {
	return &MarketplaceConfig{
		ApiKey: getEnv("WB_API_KEY", ""),
	}
}
