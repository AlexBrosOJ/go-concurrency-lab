package textstats

import (
	"strings"
	"unicode/utf8"
)

// Summary описывает число слов и рун во входной строке.
type Summary struct {
	Words int
	Runes int
}

// Analyze возвращает статистику. Слова разделяются пробельными символами.
func Analyze(text string) Summary {
	return Summary{
		Words: countWords(text),
		Runes: utf8.RuneCountInString(text),
	}
}

// countWords доступна только внутри пакета textstats.
func countWords(text string) int {
	return len(strings.Fields(text))
}
