package models

type SizeEntity struct {
	Descriptor SizeDescriptorEnum
	Type       SizeTypeEnum
	Value      float64
	Item       string
	Unit       string
}

type SizeWrapper struct {
	Descriptor SizeDescriptorEnum `json:"descriptor"`
	Type       SizeTypeEnum       `json:"type"`
	Value      float64            `json:"value"`
	Unit       string             `json:"unit"`
}

type SizeData struct {
	GlobalID int
	Sizes    []SizeWrapper
}

type SizeDescriptorEnum string
type SizeTypeEnum string

const (
	LENGTH   SizeDescriptorEnum = "LENGTH"
	DEPTH    SizeDescriptorEnum = "DEPTH"
	WIDTH    SizeDescriptorEnum = "WIDTH"
	VOLUME   SizeDescriptorEnum = "VOLUME"
	WEIGHT   SizeDescriptorEnum = "WEIGHT"
	DIAMETER SizeDescriptorEnum = "DIAMETER"
	COMMON   SizeTypeEnum       = "COMMON"
	MIN      SizeTypeEnum       = "MIN"
	MAX      SizeTypeEnum       = "MAX"
)

var DescriptorTranslationDict = map[string]SizeDescriptorEnum{
	"длина":   LENGTH,
	"глубина": DEPTH,
	"ширина":  WIDTH,
	"объем":   VOLUME,
	"вес":     WEIGHT,
	"диаметр": DIAMETER,
}

var DefaultUnits = map[SizeDescriptorEnum]string{
	LENGTH:   "cm",
	DEPTH:    "cm",
	WIDTH:    "cm",
	VOLUME:   "ml",
	WEIGHT:   "g",
	DIAMETER: "cm",
}

var UnitNormalizationMap = map[string]string{
	"м.":          "m",
	"м":           "m",
	"метров":      "m",
	"метр":        "m",
	"метры":       "m",
	"метр.":       "m",
	"см":          "cm",
	"сантиметров": "cm",
	"сантиметр":   "cm",
	"сантиметры":  "cm",
	"мм":          "mm",
	"миллиметров": "mm",
	"миллиметр":   "mm",
	"миллиметры":  "mm",
	"г":           "g",
	"гр":          "g",
	"грамм":       "g",
	"граммы":      "g",
	"кг":          "kg",
	"килограмм":   "kg",
	"килограммы":  "kg",
	"мл":          "ml",
	"миллилитров": "ml",
	"миллилитр":   "ml",
	"миллилитры":  "ml",
	"л":           "l",
	"литр":        "l",
	"литров":      "l",
	"литры":       "l",
}
var TypeTranslationDict = map[string]SizeTypeEnum{
	"общий": COMMON,
}
