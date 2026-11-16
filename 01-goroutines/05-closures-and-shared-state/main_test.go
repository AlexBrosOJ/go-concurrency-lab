package main

import (
	"slices"
	"testing"
)

func TestSquares(t *testing.T) {
	for _, values := range [][]int{nil, {}, {0}, {-3, 2, -3, 0}} {
		before := slices.Clone(values)
		got := squares(values)
		if len(got) != len(values) {
			t.Fatalf("length = %d, want %d", len(got), len(values))
		}
		for i, value := range values {
			if got[i] != value*value {
				t.Fatalf("result[%d] = %d, want %d", i, got[i], value*value)
			}
		}
		if !slices.Equal(values, before) {
			t.Fatal("input was modified")
		}
	}
}

func TestSharedCounter(t *testing.T) {
	for _, tt := range []struct{ workers, increments, want int }{
		{0, 100, 0}, {5, 0, 0}, {1, 10, 10}, {8, 1000, 8000},
	} {
		if got := sharedCounter(tt.workers, tt.increments); got != tt.want {
			t.Fatalf("sharedCounter(%d, %d) = %d, want %d", tt.workers, tt.increments, got, tt.want)
		}
	}
}
