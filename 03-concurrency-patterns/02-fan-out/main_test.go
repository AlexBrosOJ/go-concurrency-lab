package main

import (
	"slices"
	"strconv"
	"testing"
	"time"
)

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

func TestFanOut(t *testing.T) {
	tests := []struct {
		name         string
		workers      int
		values, want []int
	}{
		{name: "empty input", workers: 3},
		{name: "single worker", workers: 1, values: []int{3, 1, 2}, want: []int{1, 4, 9}},
		{name: "multiple workers", workers: 3, values: []int{1, 2, 3, 4, 5, 6}, want: []int{1, 4, 9, 16, 25, 36}},
		{name: "more workers than jobs", workers: 5, values: []int{2}, want: []int{4}},
		{name: "duplicates zero and negatives", workers: 2, values: []int{-2, 0, 2, 2}, want: []int{0, 4, 4, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs := make(chan int)
			stop := make(chan struct{})
			t.Cleanup(func() { close(stop) })
			go func() {
				defer close(jobs)
				for _, value := range tt.values {
					select {
					case jobs <- value:
					case <-stop:
						return
					}
				}
			}()
			got := collect(t, fanOut(jobs, tt.workers))
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFanOutInvalidWorkerCount(t *testing.T) {
	for _, workers := range []int{0, -1} {
		t.Run(strconv.Itoa(workers), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic for nonpositive worker count")
				}
			}()
			fanOut(nil, workers)
		})
	}
}
