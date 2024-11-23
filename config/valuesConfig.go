package config

type ValuesConfig interface {
}

type WildberriesValuesConfig struct {
	PackageHeight int `yaml:"package-height"`
	PackageWidth  int `yaml:"package-width"`
	PackageLength int `yaml:"package-length"`
}
