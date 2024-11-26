package config

import (
	"gomarketplace_api/config/values"
	"gopkg.in/yaml.v3"
	"os"
)

type Config interface {
}

type MarketplaceConfig interface {
}

type WildberriesConfig struct {
	ApiKey   string                         `yaml:"api_key"`
	WbValues values.WildberriesValues       `yaml:"default_values"`
	WbBanned values.WildberriesBannedBrands `yaml:"brands"`
}

type AppConfig struct {
	Wildberries *WildberriesConfig `yaml:"wildberries"`
	Postgres    *PostgresConfig    `yaml:"postgres"`
}

func (c *AppConfig) LoadConfig(filename string) (*AppConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	config := &AppConfig{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}
