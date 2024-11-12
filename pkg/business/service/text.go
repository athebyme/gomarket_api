package service

import (
	"bytes"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ITextService interface {
	RemoveTags(input string) string
	RemoveSpecialChars(input string) string
	RemoveAllTags(input string) string
	ReduceToLength(input string, length int) string
	ClearAndReduce(input string, length int) string
	FitIfPossible(input string, fit string, length int) string
	RemoveLinks(input string) string
	SmartReduceToLength(input string, length int) string
	RemoveUnimportantSymbols(input string) string
	ReplaceEngLettersToRus(input string) string
	RemoveWord(input string, word string) string
	ReplaceSymbols(input string, replace map[string]string) string
	AddWordIfNotExistsToFront(input string, word string) string
	AddWordIfNotExistsToEnd(input string, word string) string
	ValidateUTF8Word(word string) string
}

var engRusMap = map[string]string{
	"a": "а",
	"e": "е",
	"c": "с",
	"x": "х",
	"p": "р",
	"o": "о",
	"A": "А",
	"E": "Е",
	"P": "Р",
	"C": "С",
	"X": "Х",
	"O": "О",
	"H": "Н",
}

var prepositions = map[string]struct{}{
	"в":              {},
	"без":            {},
	"до":             {},
	"из":             {},
	"к":              {},
	"на":             {},
	"по":             {},
	"о":              {},
	"от":             {},
	"перед":          {},
	"при":            {},
	"про":            {},
	"с":              {},
	"у":              {},
	"за":             {},
	"над":            {},
	"об":             {},
	"под":            {},
	"для":            {},
	"через":          {},
	"между":          {},
	"благодаря":      {},
	"вместо":         {},
	"внутри":         {},
	"вокруг":         {},
	"вопреки":        {},
	"впереди":        {},
	"вплоть":         {},
	"вслед":          {},
	"вследствие":     {},
	"накануне":       {},
	"насчёт":         {},
	"невзирая":       {},
	"сверх":          {},
	"среди":          {},
	"сквозь":         {},
	"посредством":    {},
	"ради":           {},
	"ввиду":          {},
	"взамен":         {},
	"исключая":       {},
	"помимо":         {},
	"посередине":     {},
	"при помощи":     {},
	"с учётом":       {},
	"со стороны":     {},
	"по отношению к": {},
	"about":          {},
	"above":          {},
	"across":         {},
	"after":          {},
	"against":        {},
	"along":          {},
	"among":          {},
	"around":         {},
	"at":             {},
	"before":         {},
	"behind":         {},
	"below":          {},
	"beneath":        {},
	"beside":         {},
	"between":        {},
	"beyond":         {},
	"by":             {},
	"despite":        {},
	"down":           {},
	"during":         {},
	"except":         {},
	"for":            {},
	"from":           {},
	"in":             {},
	"inside":         {},
	"into":           {},
	"like":           {},
	"near":           {},
	"of":             {},
	"off":            {},
	"on":             {},
	"onto":           {},
	"out":            {},
	"outside":        {},
	"over":           {},
	"past":           {},
	"since":          {},
	"through":        {},
	"throughout":     {},
	"till":           {},
	"to":             {},
	"toward":         {},
	"under":          {},
	"underneath":     {},
	"until":          {},
	"up":             {},
	"upon":           {},
	"with":           {},
	"within":         {},
	"without":        {},
}

type TextService struct {
}

func NewTextService() *TextService {
	return &TextService{}
}

func (ts *TextService) ReplaceEngLettersToRus(input string) string {
	for k, v := range engRusMap {
		input = strings.Replace(input, k, v, -1)
	}
	return input
}

func (ts *TextService) ReplaceSymbols(input string, replace map[string]string) string {
	for k, v := range replace {
		input = strings.Replace(input, k, v, -1)
	}
	return input
}

func (ts *TextService) RemoveTags(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html.UnescapeString(input), "")
}

func (ts *TextService) RemoveSpecialChars(input string) string {
	var builder strings.Builder
	for _, r := range input {
		if !strings.ContainsRune("•@#$%^&*_[]{}|;'\"<>/®™▪▪️️", r) {
			builder.WriteString(string(r))
		}
	}
	return builder.String()
}

func (ts *TextService) RemoveAllTags(input string) string {
	return ts.RemoveSpecialChars(ts.RemoveTags(input))
}

func (ts *TextService) ReduceToLength(input string, length int) string {
	var builder strings.Builder
	words := strings.Split(input, " ")
	totalLength := 0
	lastNonPrepositionIndex := -1

	for i, word := range words {
		if totalLength+len(word) > length {
			break
		}

		if _, ok := prepositions[strings.ToLower(word)]; ok && totalLength+len(word)+3 >= length {
			break
		}

		if i > 0 {
			builder.WriteString(" ")
			totalLength++
		}

		builder.WriteString(word)
		totalLength += len(word)

		if _, ok := prepositions[strings.ToLower(word)]; !ok {
			lastNonPrepositionIndex = builder.Len()
		}
	}

	// избавляемся от предлога в конце
	if lastNonPrepositionIndex != -1 && lastNonPrepositionIndex < builder.Len() {
		return builder.String()[:lastNonPrepositionIndex]
	}

	return builder.String()
}
func (ts *TextService) ClearAndReduce(input string, length int) string {
	// Шаг 1: Удаляем все теги
	cleaned := ts.RemoveAllTags(input)
	// Шаг 2: Удаляем ссылки
	cleaned = ts.RemoveLinks(cleaned)
	// Шаг 3: Умное сокращение до нужной длины
	return ts.SmartReduceToLength(cleaned, length)
}
func (ts *TextService) RemoveLinks(input string) string {
	re := regexp.MustCompile(`https?://[^\s]+`)
	return re.ReplaceAllString(input, "")
}

func (ts *TextService) SmartReduceToLength(input string, length int) string {
	cleaned := input
	if len(cleaned) > length {
		cleaned = ts.RemoveUnimportantSymbols(input)
	}
	return ts.ReduceToLength(cleaned, length)
}

func (ts *TextService) RemoveUnimportantSymbols(input string) string {
	re := regexp.MustCompile(`[%№(),."'|/\-+&]`)
	return re.ReplaceAllString(input, "")
}

func (ts *TextService) RemoveWord(input string, word string) string {
	return regexp.MustCompile(word).ReplaceAllString(input, "")
}

func (ts *TextService) AddWordIfNotExistsToFront(input string, word string) string {
	return ts.AddWordIfNotExists(input, word, 0)
}

func (ts *TextService) AddWordIfNotExistsToEnd(input string, word string) string {
	return ts.AddWordIfNotExists(input, word, len(input))
}

func (ts *TextService) AddWordIfNotExists(input string, word string, index int) string {
	word = ts.ValidateUTF8Word(word)

	safeWord := regexp.QuoteMeta(ts.TrimLastRunes(word, 3))

	match := regexp.MustCompile(safeWord).Find([]byte(input))
	if match == nil {
		if index > len(input) {
			index = len(input)
		}
		return input[:index] + word + " " + input[index:]
	}
	return input
}

func (ts *TextService) FitIfPossible(input string, fit string, length int) string {
	if len(input)+len(fit)+1 > length {
		return input
	}
	return ts.AddWordIfNotExistsToEnd(input, fit)
}

func (ts *TextService) AddCategoryIfNotExistInAppellation(appellation string, category string) string {
	// Проверка UTF-8 корректности для appellation и category
	appellation = ts.ValidateUTF8Word(appellation)
	category = ts.ValidateUTF8Word(category)

	// Разделяем category на отдельные слова
	words := strings.Fields(category)

	// Создаем строку для слов, которых нет в appellation
	newWords := ""

	for _, word := range words {
		// Если слово отсутствует в appellation, добавляем его в newWords
		if ts.AddWordIfNotExists(appellation, word, 0) == appellation {
			newWords += word + " "
		}
	}

	// Объединяем newWords и appellation
	return strings.TrimSpace(newWords) + " " + appellation
}

func (ts *TextService) ValidateUTF8Word(word string) string {
	if !utf8.ValidString(word) {
		return string(bytes.ReplaceAll([]byte(word), []byte{0xEF, 0xBF, 0xBD}, []byte("?")))
	}
	return word
}
func (ts *TextService) TrimLastRunes(word string, n int) string {
	runes := []rune(word)
	if len(runes) >= n {
		return string(runes[:len(runes)-n])
	}
	return word
}
