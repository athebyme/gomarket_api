package service

import (
	"regexp"
	"strings"
	"unicode"
)

type ITextService interface {
	RemoveTags(input string) string
	RemoveSpecialChars(input string) string
	RemoveAllTags(input string) string
}

type TextServiceImpl struct {
}

func NewTextService() ITextService {
	return &TextServiceImpl{}
}

func (impl *TextServiceImpl) RemoveTags(input string) string {
	// Регулярные выражения для удаления PHP, JS, HTML тегов и спецсимволов
	phpTagPattern := regexp.MustCompile(`<\?php[\s\S]*?\?>`)
	jsTagPattern := regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	htmlTagPattern := regexp.MustCompile(`<.*?>`)
	alphaNumericPattern := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)

	// Создаем StringBuilder для накопления результата
	var builder strings.Builder

	// Удаляем PHP теги и содержимое
	input = phpTagPattern.ReplaceAllString(input, "")
	// Удаляем JavaScript теги и содержимое
	input = jsTagPattern.ReplaceAllString(input, "")
	// Удаляем HTML теги
	input = htmlTagPattern.ReplaceAllString(input, "")
	// Удаляем спецсимволы
	input = alphaNumericPattern.ReplaceAllString(input, "")

	// Добавляем очищенный результат в builder
	builder.WriteString(input)

	return builder.String()
}

func (impl *TextServiceImpl) RemoveSpecialChars(input string) string {
	panic("to do")
}

func (impl *TextServiceImpl) RemoveAllTags(input string) string {
	phpTagPattern := regexp.MustCompile(`<\?php[\s\S]*?\?>`)
	jsTagPattern := regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	htmlTagPattern := regexp.MustCompile(`<.*?>`)

	input = phpTagPattern.ReplaceAllString(input, "")
	input = jsTagPattern.ReplaceAllString(input, "")
	input = htmlTagPattern.ReplaceAllString(input, "")

	var builder strings.Builder

	for _, r := range []rune(input) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}
