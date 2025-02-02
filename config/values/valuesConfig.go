package values

type Config interface {
}

type WildberriesValues struct {
	PackageHeight int `yaml:"package-height"`
	PackageWidth  int `yaml:"package-width"`
	PackageLength int `yaml:"package-length"`
}

type WildberriesBannedBrands struct {
	BannedBrands []string `yaml:"banned"`
}

type Identity struct {
	Code int `yaml:"code"`
}
