package main

import (
	"slices"
	"testing"
	"time"
)

const testTimeout = 2 * time.Second

// Один таймер ограничивает весь сбор, включая ожидание close.
func collect(t *testing.T, in <-chan int) []int {
	t.Helper()
	// При раннем выходе продолжаем читать, освобождая производителей.
	t.Cleanup(func() {
		timer := time.NewTimer(testTimeout)
		defer timer.Stop()
		for {
			select {
			case _, ok := <-in:
				if !ok {
					return
				}
			case <-timer.C:
				t.Error("stream did not close during cleanup")
				return
			}
		}
	})
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	var values []int
	for {
		select {
		case value, ok := <-in:
			if !ok {
				return values
			}
			values = append(values, value)
		case <-timer.C:
			t.Fatal("timed out waiting for values and stream closure")
			return nil
		}
	}
}

func TestStream(t *testing.T) {
	tests := []struct {
		name   string
		values []int
	}{
		{name: "nil input"},
		{name: "empty input", values: []int{}},
		{name: "single value", values: []int{42}},
		{name: "order zero negatives duplicates", values: []int{3, 0, -2, 3, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collect(t, stream(tt.values)); !slices.Equal(got, tt.values) {
				t.Fatalf("received %v, want %v in order", got, tt.values)
			}
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		want   int
	}{
		{name: "closed empty"},
		{name: "positive", values: []int{1, 2, 3, 4}, want: 10},
		{name: "zero is data", values: []int{0, 5}, want: 5},
		{name: "negative and duplicates", values: []int{-3, 2, 2}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := make(chan int, len(tt.values))
			for _, value := range tt.values {
				in <- value
			}
			close(in)
			result := make(chan int, 1)
			go func() { result <- sum(in) }()
			timer := time.NewTimer(testTimeout)
			defer timer.Stop()
			select {
			case got := <-result:
				if got != tt.want {
					t.Fatalf("sum = %d, want %d", got, tt.want)
				}
			case <-timer.C:
				t.Fatal("range did not finish on a closed channel")
			}
		})
	}
}

func TestSharedStream(t *testing.T) {
	tests := []struct {
		name   string
		groups [][]int
		want   []int
	}{
		{name: "no producers"},
		{name: "empty producers", groups: [][]int{{}, {}}},
		{name: "one producer", groups: [][]int{{3, 1}}, want: []int{1, 3}},
		{name: "unequal lengths and empty group", groups: [][]int{{1, 3, 5}, {}, {2}}, want: []int{1, 2, 3, 5}},
		{name: "duplicates and zero", groups: [][]int{{0, -1}, {0, 2, -1}}, want: []int{-1, -1, 0, 0, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collect(t, sharedStream(tt.groups...))
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("received %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClosedChannelDistinguishesZeroFromEnd(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 0
	close(ch)
	// collect ограничивает ожидание таймером и учитывает ok.
	if got := collect(t, ch); !slices.Equal(got, []int{0}) {
		t.Fatalf("received %v, want one zero before channel end", got)
	}
	for i := 0; i < 2; i++ {
		timer := time.NewTimer(testTimeout)
		select {
		case value, ok := <-ch:
			timer.Stop()
			if value != 0 || ok {
				t.Fatalf("received (%d, %t), want (0, false)", value, ok)
			}
		case <-timer.C:
			t.Fatal("closed drained channel blocked")
		}
	}
}

func TestInvalidChannelOperations(t *testing.T) {
	tests := []struct {
		name   string
		action func()
	}{
		{name: "send after close", action: func() { ch := make(chan int, 1); close(ch); ch <- 1 }},
		{name: "close twice", action: func() { ch := make(chan int); close(ch); close(ch) }},
		{name: "close nil", action: func() { var ch chan int; close(ch) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			tt.action()
		})
	}
}
