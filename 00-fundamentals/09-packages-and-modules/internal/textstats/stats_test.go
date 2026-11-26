package textstats_test

import (
	"testing"

	"go-concurrency-lab/00-fundamentals/09-packages-and-modules/internal/textstats"
)

// Тест использует только функции, доступные вызывающему пакету.
func TestAnalyze(t *testing.T) {
	for _, test := range []struct {
		input string
		want  textstats.Summary
	}{
		{"", textstats.Summary{}},
		{"  ", textstats.Summary{Runes: 2}},
		{"ёж и кот", textstats.Summary{Words: 3, Runes: 8}},
		{"Go\u00a0lab", textstats.Summary{Words: 2, Runes: 6}},
	} {
		if got := textstats.Analyze(test.input); got != test.want {
			t.Fatalf("Analyze(%q) = %+v, want %+v", test.input, got, test.want)
		}
	}
}
