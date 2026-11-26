package main

import (
	"math"
	"slices"
	"testing"
)

func TestFirstDistinguishesZeroFromEmpty(t *testing.T) {
	if got, ok := first([]int{0}); got != 0 || !ok {
		t.Fatalf("got (%d, %t)", got, ok)
	}
	if got, ok := first[string](nil); got != "" || ok {
		t.Fatalf("got (%q, %t)", got, ok)
	}
}

func TestContainsWithComparableStruct(t *testing.T) {
	type label struct{ name string }
	values := []label{{name: "Go"}, {name: "SQL"}}
	if !contains(values, label{name: "SQL"}) || contains(values, label{}) || contains([]label(nil), label{}) {
		t.Fatal("unexpected membership result")
	}
}

func TestMinValue(t *testing.T) {
	type score int
	if got, ok := minValue([]score{5, -3, -3, 0}); got != -3 || !ok {
		t.Fatalf("got (%d, %t)", got, ok)
	}
	if got, ok := minValue([]float64{2, math.NaN(), -5}); !ok || !math.IsNaN(got) {
		t.Fatalf("got (%v, %t), want NaN under cmp.Less order", got, ok)
	}
	if got, ok := minValue([]int(nil)); got != 0 || ok {
		t.Fatalf("got (%d, %t)", got, ok)
	}
}

func TestReversedPreservesTypeAndUsesSeparateStorage(t *testing.T) {
	original := Scores{1, 2, 3}
	got := reversed(original)
	if !slices.Equal(got, Scores{3, 2, 1}) || got.Count() != 3 {
		t.Fatalf("unexpected result: %v", got)
	}
	got[0] = 999
	if !slices.Equal(original, Scores{1, 2, 3}) {
		t.Fatalf("original changed: %v", original)
	}
	if reversed(Scores(nil)) != nil || reversed(Scores{}) == nil {
		t.Fatal("nil and empty slices were not preserved")
	}
}
