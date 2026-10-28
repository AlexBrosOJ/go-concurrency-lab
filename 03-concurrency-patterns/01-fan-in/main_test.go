package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestFanIn(t *testing.T) {
	first := make(chan int, 3)
	second := make(chan int, 3)

	for _, value := range []int{1, 3, 5} {
		first <- value
	}
	for _, value := range []int{2, 4, 6} {
		second <- value
	}

	close(first)
	close(second)

	var got []int
	for value := range fanIn(first, second) {
		got = append(got, value)
	}

	sort.Ints(got)
	want := []int{1, 2, 3, 4, 5, 6}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
