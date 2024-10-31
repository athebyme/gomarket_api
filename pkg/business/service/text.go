package service

import (
	"html"
	"regexp"
	"strings"
)

type ITextService interface {
	RemoveTags(input string) string
	RemoveSpecialChars(input string) string
	RemoveAllTags(input string) string
	ReduceToLength(input string, length int) string
	ClearAndReduce(input string, length int) string
	RemoveLinks(input string) string
	SmartReduceToLength(input string, length int) string
	RemoveUnimportantSymbols(input string) string
}

type TextService struct{}

func NewTextService() *TextService {
	return &TextService{}
}

func (ts *TextService) RemoveTags(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html.UnescapeString(input), "")
}

func (ts *TextService) RemoveSpecialChars(input string) string {
	var builder strings.Builder
	for _, r := range input {
		if !strings.ContainsRune("•@#$%^&*_[]{}|;'\"<>/", r) {
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

	for i, word := range words {
		if totalLength+len(word) > length {
			break
		}

		if i > 0 {
			builder.WriteString(" ")
			totalLength++
		}

		builder.WriteString(word)
		totalLength += len(word)
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
	re := regexp.MustCompile(`[(),."'|/\-+&]`)
	return re.ReplaceAllString(input, "")
}
