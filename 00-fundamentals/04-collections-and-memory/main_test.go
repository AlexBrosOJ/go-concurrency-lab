package main

import (
	"slices"
	"testing"
)

func TestCloneAndAppend(t *testing.T) {
	spare := make([]int, 2, 5)
	copy(spare, []int{10, 20})
	for _, tt := range []struct {
		name   string
		source []int
		value  int
		want   []int
	}{
		{name: "nil", value: 3, want: []int{3}},
		{name: "empty", source: []int{}, value: 0, want: []int{0}},
		{name: "full capacity", source: []int{1, -1}, value: 2, want: []int{1, -1, 2}},
		{name: "spare capacity", source: spare, value: 30, want: []int{10, 20, 30}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			before := slices.Clone(tt.source)
			// Сохраняем и невидимую за len часть доступного массива.
			backing := tt.source[:cap(tt.source)]
			beforeBacking := slices.Clone(backing)
			got := cloneAndAppend(tt.source, tt.value)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			if !slices.Equal(backing, beforeBacking) {
				t.Fatal("append modified source backing array")
			}
			got[0] = 999
			if !slices.Equal(tt.source, before) {
				t.Fatal("result shares mutable elements with source")
			}
			if len(tt.source) > 0 {
				tt.source[0] = -999
				if got[0] != 999 {
					t.Fatal("source modification affected result")
				}
			}
		})
	}
}
