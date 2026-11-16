package main

import (
	"slices"
	"testing"
)

func closedInput(values ...int) <-chan int {
	ch := make(chan int, len(values))
	for _, value := range values {
		ch <- value
	}
	close(ch)
	return ch
}

func TestDrainTwo(t *testing.T) {
	for _, tt := range []struct {
		name          string
		first, second <-chan int
		want          []int
	}{
		{name: "both nil"},
		{name: "both closed empty", first: closedInput(), second: closedInput()},
		{name: "first nil", second: closedInput(0, 2, 2), want: []int{0, 2, 2}},
		{name: "second nil", first: closedInput(1, 3), want: []int{1, 3}},
		{name: "closed and active", first: closedInput(), second: closedInput(4), want: []int{4}},
		{name: "both populated", first: closedInput(1, 2), second: closedInput(10, 20), want: []int{1, 2, 10, 20}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := drainTwo(tt.first, tt.second)
			// Порядок между входами не определён; сравниваем мультимножества.
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDrainTwoConcurrentProducersPreserveInputOrder(t *testing.T) {
	first, second := make(chan int), make(chan int)
	go func() {
		defer close(first)
		for _, value := range []int{1, 2, 3} {
			first <- value
		}
	}()
	go func() {
		defer close(second)
		for _, value := range []int{10, 20, 30} {
			second <- value
		}
	}()
	got := drainTwo(first, second)
	var small, large []int
	for _, value := range got {
		if value < 10 {
			small = append(small, value)
		} else {
			large = append(large, value)
		}
	}
	if !slices.Equal(small, []int{1, 2, 3}) || !slices.Equal(large, []int{10, 20, 30}) {
		t.Fatalf("lost values or input order: %v", got)
	}
}
