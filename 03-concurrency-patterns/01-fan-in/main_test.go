package main

import (
	"slices"
	"testing"
	"time"
)

// Один таймер ограничивает весь сбор, включая ожидание закрытия канала.
func collect(t *testing.T, ch <-chan int) []int {
	t.Helper()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	var values []int
	for {
		select {
		case value, ok := <-ch:
			if !ok {
				return values
			}
			values = append(values, value)
		case <-timer.C:
			t.Fatal("timed out waiting for values and channel closure")
			return nil
		}
	}
}

func TestFanIn(t *testing.T) {
	tests := []struct {
		name   string
		inputs [][]int
		want   []int
	}{
		{name: "no inputs"},
		{name: "one empty input", inputs: [][]int{{}}},
		{name: "several empty inputs", inputs: [][]int{{}, {}}},
		{name: "single input", inputs: [][]int{{3, 1, 2}}, want: []int{1, 2, 3}},
		{name: "multiple inputs", inputs: [][]int{{1, 3, 5}, {2, 4, 6}}, want: []int{1, 2, 3, 4, 5, 6}},
		{name: "empty and nonempty", inputs: [][]int{{}, {2, 1}, {}}, want: []int{1, 2}},
		{name: "duplicates and zero", inputs: [][]int{{0, -1, 2}, {2, -1}}, want: []int{-1, -1, 0, 2, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputs []<-chan int
			for _, values := range tt.inputs {
				ch := make(chan int, len(values))
				for _, value := range values {
					ch <- value
				}
				close(ch)
				inputs = append(inputs, ch)
			}
			got := collect(t, fanIn(inputs...))
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFanInUnbufferedInputs(t *testing.T) {
	first := make(chan int)
	second := make(chan int)
	stop := make(chan struct{})
	t.Cleanup(func() { close(stop) })
	// Каждый производитель закрывает свой вход независимо от другого.
	for ch, values := range map[chan int][]int{first: {1, 3, 5}, second: {2}} {
		go func() {
			defer close(ch)
			for _, value := range values {
				select {
				case ch <- value:
				case <-stop:
					return
				}
			}
		}()
	}
	got := collect(t, fanIn(first, second))
	slices.Sort(got)
	if want := []int{1, 2, 3, 5}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFanInPreservesSingleInputOrder(t *testing.T) {
	ch := make(chan int, 3)
	for _, value := range []int{3, 1, 2} {
		ch <- value
	}
	close(ch)
	if got, want := collect(t, fanIn(ch)), []int{3, 1, 2}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
