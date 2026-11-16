package main

import (
	"errors"
	"testing"
)

func TestDivide(t *testing.T) {
	for _, tt := range []struct {
		name       string
		a, b, want float64
		wantErr    bool
	}{
		{name: "fraction", a: 10, b: 4, want: 2.5},
		{name: "negative", a: -9, b: 3, want: -3},
		{name: "zero numerator", b: 3},
		{name: "zero denominator", a: 5, wantErr: true},
		{name: "both zero", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := divide(tt.a, tt.b)
			if tt.wantErr {
				if !errors.Is(err, errDivisionByZero) {
					t.Fatalf("got %v, want division error", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got (%v, %v), want (%v, nil)", got, err, tt.want)
			}
		})
	}
}

func TestSum(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values []int
		want   int
	}{
		{name: "nil"}, {name: "empty", values: []int{}},
		{name: "positive", values: []int{1, 2, 3}, want: 6},
		{name: "mixed and repeated", values: []int{3, -2, 0, 3}, want: 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := sum(tt.values...); got != tt.want {
				t.Fatalf("sum = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValueAndPointerParameters(t *testing.T) {
	value := 10
	changeCopy(value)
	if value != 10 {
		t.Fatal("value parameter changed the caller's variable")
	}
	changeOriginal(&value)
	if value != 99 {
		t.Fatalf("pointer update = %d, want 99", value)
	}
}

func TestCountersKeepIndependentState(t *testing.T) {
	first, second := newCounter(), newCounter()
	for _, want := range []int{1, 2, 3} {
		if got := first(); got != want {
			t.Fatalf("first = %d, want %d", got, want)
		}
	}
	if got := second(); got != 1 {
		t.Fatalf("second = %d, want 1", got)
	}
	if got := first(); got != 4 {
		t.Fatalf("first = %d, want 4", got)
	}
}

func TestApply(t *testing.T) {
	calls := 0
	got := apply(7, func(value int) int { calls++; return -value })
	if got != -7 || calls != 1 {
		t.Fatalf("got %d and %d calls, want -7 and 1", got, calls)
	}
}
