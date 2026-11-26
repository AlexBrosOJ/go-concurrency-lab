package main

import (
	"context"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

type collected struct {
	values []int
	err    error
}

func collectAsync(ctx context.Context, timeout time.Duration, input <-chan int) <-chan collected {
	result := make(chan collected, 1)
	go func() {
		values, err := collectWithin(ctx, timeout, input)
		result <- collected{values, err}
	}()
	return result
}

func awaitCollected(t *testing.T, result <-chan collected) collected {
	t.Helper()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case got := <-result:
		return got
	case <-timer.C:
		t.Fatal("collector did not return")
		return collected{}
	}
}

func TestCollectWithinFinishesOnClose(t *testing.T) {
	for _, want := range [][]int{nil, {4, 0, 4, -7}} {
		ctx, cancel := context.WithCancel(context.Background())
		got := awaitCollected(t, collectAsync(ctx, time.Hour, closedInput(want...)))
		cancel()
		if got.err != nil || !slices.Equal(got.values, want) {
			t.Fatalf("got %+v, want %v with no error", got, want)
		}
	}
}

func TestCollectWithinKeepsOneBudgetAndPartialResults(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		input := make(chan int, 1)
		input <- 0
		start := time.Now()
		result := collectAsync(context.Background(), time.Second, input)
		<-time.After(750 * time.Millisecond)
		input <- 9
		got := awaitCollected(t, result)
		if got.err != context.DeadlineExceeded || !slices.Equal(got.values, []int{0, 9}) {
			t.Fatalf("got %+v, want both values and deadline exceeded", got)
		}
		// Второе число пришло почти в конце бюджета и не перезапустило таймер.
		if elapsed := time.Since(start); elapsed != time.Second {
			t.Fatalf("elapsed %v, want one virtual second", elapsed)
		}
	})
}

func TestCollectWithinInheritsEarlierDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		start := time.Now()
		got := awaitCollected(t, collectAsync(ctx, time.Hour, nil))
		if got.err != context.DeadlineExceeded || time.Since(start) != time.Second {
			t.Fatalf("got %+v after %v, want parent deadline", got, time.Since(start))
		}
	})
}

func TestCollectWithinStopsOnParentCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		input := make(chan int, 1)
		input <- 0
		result := collectAsync(ctx, time.Hour, input)
		synctest.Wait()
		cancel()
		got := awaitCollected(t, result)
		if got.err != context.Canceled || !slices.Equal(got.values, []int{0}) {
			t.Fatalf("got %+v, want received zero and context.Canceled", got)
		}
	})
}

func TestCollectWithinRejectsExpiredBudgetBeforeReading(t *testing.T) {
	for _, budget := range []time.Duration{0, -time.Second} {
		got := awaitCollected(t, collectAsync(context.Background(), budget, closedInput(42)))
		if got.err != context.DeadlineExceeded || len(got.values) != 0 {
			t.Fatalf("got %+v, want expired budget without data", got)
		}
	}
}
