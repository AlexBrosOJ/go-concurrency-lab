package main

import (
	"strconv"
	"testing"
)

func TestGradeBoundaries(t *testing.T) {
	for _, tt := range []struct {
		score int
		want  string
	}{
		{0, "F"}, {59, "F"}, {60, "C"}, {74, "C"},
		{75, "B"}, {89, "B"}, {90, "A"}, {100, "A"},
	} {
		t.Run(strconv.Itoa(tt.score), func(t *testing.T) {
			if got := grade(tt.score); got != tt.want {
				t.Fatalf("grade = %q, want %q", got, tt.want)
			}
		})
	}
}
