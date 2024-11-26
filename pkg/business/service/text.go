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
	EnsureSpaceAfterPeriod(input string) string
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
	// Русские предлоги
	"в": {}, "во": {}, "на": {}, "с": {}, "со": {},
	"по": {}, "о": {}, "об": {}, "обо": {}, "за": {},
	"из": {}, "изо": {}, "к": {}, "ко": {}, "от": {},
	"ото": {}, "до": {}, "у": {}, "перед": {}, "передо": {},
	"для": {}, "про": {}, "над": {}, "под": {}, "при": {},
	"через": {}, "между": {}, "без": {}, "вне": {}, "около": {},
	"возле": {}, "рядом": {}, "близ": {}, "после": {}, "посредством": {},
	"вдоль": {}, "вокруг": {}, "посредине": {}, "против": {}, "из-за": {},
	"из-под": {}, "наподобие": {}, "взамен": {}, "вместо": {}, "среди": {},
	"сквозь": {}, "благодаря": {}, "несмотря": {}, "вопреки": {},

	// Русские союзы
	"и": {}, "а": {}, "но": {}, "да": {}, "либо": {},
	"или": {}, "что": {}, "чтобы": {}, "как": {}, "когда": {},
	"если": {}, "хотя": {}, "потому": {}, "также": {}, "тоже": {},
	"однако": {}, "зато": {}, "поскольку": {}, "так": {}, "будто": {},
	"словно": {}, "только": {}, "раз": {}, "дабы": {}, "иначе": {},

	// Русские частицы
	"бы": {}, "ли": {}, "же": {}, "пусть": {}, "ведь": {},
	"уж": {}, "лишь": {}, "даже": {}, "вот": {}, "вон": {},
	"разве": {}, "едва ли": {}, "точно": {}, "неужели": {}, "именно": {},
	"все-таки": {}, "например": {}, "как раз": {}, "только лишь": {},

	// Английские предлоги
	"at": {}, "by": {}, "for": {}, "from": {}, "in": {},
	"of": {}, "on": {}, "to": {}, "with": {}, "about": {},
	"against": {}, "between": {}, "into": {}, "through": {}, "during": {},
	"before": {}, "after": {}, "above": {}, "below": {}, "under": {},
	"over": {}, "around": {}, "among": {}, "across": {}, "behind": {},
	"beyond": {}, "beside": {}, "near": {}, "outside": {}, "inside": {},
	"without": {}, "within": {}, "along": {}, "towards": {}, "upon": {},

	// Английские союзы
	"and": {}, "or": {}, "but": {}, "nor": {},
	"so": {}, "although": {}, "because": {}, "since": {},
	"unless": {}, "while": {}, "whereas": {}, "if": {}, "then": {},
	"however": {}, "whether": {}, "either": {}, "neither": {}, "as": {},
	"though": {}, "until": {}, "once": {}, "even if": {}, "rather than": {},

	// Английские частицы
	"not": {}, "no": {}, "yes": {}, "indeed": {}, "only": {},
	"just": {}, "almost": {}, "also": {}, "even": {}, "still": {},
	"yet": {}, "too": {},
	"perhaps": {}, "maybe": {}, "instead": {}, "thus": {},
	"moreover": {}, "therefore": {}, "nonetheless": {}, "meanwhile": {},
}

type TextService struct {
}

func NewTextService() *TextService {
	return &TextService{}
}

func (ts *TextService) ReplaceEngLettersToRus(input string) string {
	for k, v := range engRusMap {
		if strings.Contains(input, k) {
			return strings.Replace(input, k, v, 1)
		}
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
	cleaned = ts.SmartReduceToLength(cleaned, length)
	// Шаг 4: Убедиться, что после точки есть пробел
	return ts.EnsureSpaceAfterPeriod(cleaned)
}

func (ts *TextService) RemoveLinks(input string) string {
	// Регулярное выражение для удаления ссылок вида "https://", "http://", и просто доменов
	re := regexp.MustCompile(`https?:?/?/?[^\s]+|(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}`)
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
func (ts *TextService) EnsureSpaceAfterPeriod(input string) string {
	re := regexp.MustCompile(`\.(\S)`)
	return re.ReplaceAllString(input, ". $1")
}
