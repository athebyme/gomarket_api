package storage

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Size struct {
	Descriptor SizeDescriptorEnum
	Type       SizeTypeEnum
	Value      float64
	Item       string
	Unit       string
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

var descriptorTranslationDict = map[string]SizeDescriptorEnum{
	"длина":   LENGTH,
	"глубина": DEPTH,
	"ширина":  WIDTH,
	"объем":   VOLUME,
	"вес":     WEIGHT,
	"диаметр": DIAMETER,
}

var defaultUnits = map[SizeDescriptorEnum]string{
	LENGTH:   "cm",
	DEPTH:    "cm",
	WIDTH:    "cm",
	VOLUME:   "ml",
	WEIGHT:   "g",
	DIAMETER: "cm",
}

var unitNormalizationMap = map[string]string{
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
var typeTranslationDict = map[string]SizeTypeEnum{
	"общий": COMMON,
}

func ParseSizes(text string) ([]Size, error) {
	result := []Size{}

	for word, descriptor := range descriptorTranslationDict {
		patternStr := `(?i)` + word + `(?:.*?)(\d+[.,]?\d*)(?:\s*(?:-|до|–)\s*(\d+[.,]?\d*))?\s*([a-zA-Zа-яА-Я]+)`
		re := regexp.MustCompile(patternStr)

		matches := re.FindAllStringSubmatch(text, -1)
		sizeMap := make(map[string]bool)
		for _, match := range matches {
			if len(match) < 2 { // Проверка на успешное совпадение
				continue
			}

			value1, err := strconv.ParseFloat(strings.Replace(match[1], ",", ".", -1), 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse value 1: %w", err) // Более информативное сообщение об ошибке
			}

			rawUnit := ""
			if len(match) > 3 {
				rawUnit = match[3]
			}
			unit := normalizeUnit(rawUnit, unitNormalizationMap, descriptorTranslationDict[word])

			value2 := value1 // Значение по умолчанию, если диапазона нет
			if len(match) > 2 && match[2] != "" {
				value2, err = strconv.ParseFloat(strings.Replace(match[2], ",", ".", -1), 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse value 2: %w", err) // Более информативное сообщение об ошибке
				}

			}

			// Упрощаем логику определения "item"
			item := "общий"                                                                // Значение по умолчанию
			wordsAfterKeyword := strings.TrimSpace(strings.Replace(match[0], word, "", 1)) //обрезать слово для поиска типа по оставшейся строке. типа "до кольца", "общий" и т.д.

			if wordsAfterKeyword != "" {
				_, found := typeTranslationDict[wordsAfterKeyword]
				if found {
					item = wordsAfterKeyword
				}
			}

			if value1 != value2 {
				key1 := fmt.Sprintf("%s-%s-%f", descriptor, unit, value1)
				key2 := fmt.Sprintf("%s-%s-%f", descriptor, unit, value2)

				if _, duplicate := sizeMap[key1]; !duplicate {
					result = append(result, Size{Descriptor: descriptor, Type: MIN, Value: value1, Item: item, Unit: unit})
					sizeMap[key1] = true
				}
				if _, duplicate := sizeMap[key2]; !duplicate {
					result = append(result, Size{Descriptor: descriptor, Type: MAX, Value: value2, Item: item, Unit: unit})
					sizeMap[key2] = true
				}

				sizeMap[key1] = true
				sizeMap[key2] = true
			}

			parsedSize := Size{
				Descriptor: descriptor,
				Type:       typeTranslationDict[item], /* используем item */
				Value:      value1,
				Item:       item, /* используем item */
				Unit:       unit,
			}

			key := fmt.Sprintf("%s-%s-%f", descriptor, unit, value1)
			if _, duplicate := sizeMap[key]; !duplicate {
				result = append(result, parsedSize)
				sizeMap[key] = true // Отмечаем размер как найденный
			} else {
				// Дополнительная отладка, если нужно
				log.Printf("Skipping duplicate size: %+v, key: %s, item: %s result: %v\n", parsedSize, key, item, result)
			}
		}
	}

	return result, nil
}

func normalizeUnit(rawUnit string, unitNormalizationMap map[string]string, descriptor SizeDescriptorEnum) string {
	normalizedUnit, ok := unitNormalizationMap[strings.ToLower(rawUnit)]
	if ok {
		return normalizedUnit
	}
	if defaultUnit, ok := defaultUnits[descriptor]; ok {
		return defaultUnit
	}

	return ""
}
