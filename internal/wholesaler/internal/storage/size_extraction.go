package storage

import (
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"regexp"
	"strconv"
	"strings"
)

func ParseSizes(text string) ([]models.SizeEntity, error) {
	result := []models.SizeEntity{}

	for word, descriptor := range models.DescriptorTranslationDict {
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
			unit := normalizeUnit(rawUnit, models.UnitNormalizationMap, models.DescriptorTranslationDict[word])

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
				_, found := models.TypeTranslationDict[wordsAfterKeyword]
				if found {
					item = wordsAfterKeyword
				}
			}

			if value1 != value2 {
				key1 := fmt.Sprintf("%s-%s-%f", descriptor, unit, value1)
				key2 := fmt.Sprintf("%s-%s-%f", descriptor, unit, value2)

				if _, duplicate := sizeMap[key1]; !duplicate {
					result = append(result, models.SizeEntity{Descriptor: descriptor, Type: models.MIN, Value: value1, Item: item, Unit: unit})
					sizeMap[key1] = true
				}
				if _, duplicate := sizeMap[key2]; !duplicate {
					result = append(result, models.SizeEntity{Descriptor: descriptor, Type: models.MAX, Value: value2, Item: item, Unit: unit})
					sizeMap[key2] = true
				}

				sizeMap[key1] = true
				sizeMap[key2] = true
			}

			parsedSize := models.SizeEntity{
				Descriptor: descriptor,
				Type:       models.TypeTranslationDict[item], /* используем item */
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
				continue
			}
		}
	}

	return result, nil
}

func normalizeUnit(rawUnit string, unitNormalizationMap map[string]string, descriptor models.SizeDescriptorEnum) string {
	normalizedUnit, ok := unitNormalizationMap[strings.ToLower(rawUnit)]
	if ok {
		return normalizedUnit
	}
	if defaultUnit, ok := models.DefaultUnits[descriptor]; ok {
		return defaultUnit
	}

	return ""
}
