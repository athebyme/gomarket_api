package service

import (
	"strings"
)

type ITextService interface {
	RemoveTags(input string) string
	RemoveSpecialChars(input string) string
	RemoveAllTags(input string) string
	ReduceToLength(input string, length int) string
	ClearAndReduce(input string, length int) string
}

type TextService struct{}

func NewTextService() *TextService {
	return &TextService{}
}

func (ts *TextService) RemoveTags(input string) string {
	//// Регулярные выражения для удаления PHP, JS, HTML тегов и спецсимволов
	//phpTagPattern := regexp.MustCompile(`<\?php[\s\S]*?\?>`)
	//jsTagPattern := regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	//htmlTagPattern := regexp.MustCompile(`<.*?>`)
	//alphaNumericPattern := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	//
	//// Создаем StringBuilder для накопления результата
	//var builder strings.Builder
	//
	//// Удаляем PHP теги и содержимое
	//input = phpTagPattern.ReplaceAllString(input, "")
	//// Удаляем JavaScript теги и содержимое
	//input = jsTagPattern.ReplaceAllString(input, "")
	//// Удаляем HTML теги
	//input = htmlTagPattern.ReplaceAllString(input, "")
	//// Удаляем спецсимволы
	//input = alphaNumericPattern.ReplaceAllString(input, "")
	//
	//// Добавляем очищенный результат в builder
	//builder.WriteString(input)
	//
	//return builder.String()
	return strings.ReplaceAll(input, "<[^>]*>", "")
}

func (ts *TextService) RemoveSpecialChars(input string) string {
	var builder strings.Builder
	for _, r := range input {
		if !strings.ContainsRune("@#$%^&*_[]{}|;'\"<>/", r) {
			builder.WriteString(string(r))
		}
	}
	return builder.String()
}

func (ts *TextService) RemoveAllTags(input string) string {
	//phpTagPattern := regexp.MustCompile(`<\?php[\s\S]*?\?>`)
	//jsTagPattern := regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	//htmlTagPattern := regexp.MustCompile(`<.*?>`)
	//
	//input = phpTagPattern.ReplaceAllString(input, "")
	//input = jsTagPattern.ReplaceAllString(input, "")
	//input = htmlTagPattern.ReplaceAllString(input, "")
	//
	//var builder strings.Builder
	//
	//for _, r := range []rune(input) {
	//	if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
	//		builder.WriteRune(r)
	//	}
	//}
	//
	//return builder.String()

	return ts.RemoveSpecialChars(ts.RemoveTags(input))
}

func (ts *TextService) ReduceToLength(input string, length int) string {
	var builder strings.Builder
	words := strings.Split(input, " ")
	for i, word := range words {
		if i < length-1 {
			builder.WriteString(word + " ")
		} else if i == length-1 && strings.HasSuffix(word, ",") {
			builder.WriteString(word)
		} else {
			break
		}
	}
	return builder.String()
}

func (ts *TextService) ClearAndReduce(input string, length int) string {
	return ts.ReduceToLength(ts.RemoveAllTags(input), length)
}
