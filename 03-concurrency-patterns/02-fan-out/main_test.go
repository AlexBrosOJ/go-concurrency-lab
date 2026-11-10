package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestFanOut(t *testing.T) {
	in := make(chan int, 6)
	for _, value := range []int{1, 2, 3, 4, 5, 6} {
		in <- value
	}
	close(in)

	var got []int
	for result := range fanOut(in, 3, func(value int) int {
		return value * 2
	}) {
		got = append(got, result)
	}

	sort.Ints(got)
	want := []int{2, 4, 6, 8, 10, 12}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFanOutEmptyInput(t *testing.T) {
	in := make(chan int)
	close(in)

	count := 0
	for range fanOut(in, 2, func(value int) int {
		return value
	}) {
		count++
	}

	if count != 0 {
		t.Fatalf("expected closed empty output channel, got %d values", count)
	}
}
