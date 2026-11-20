package main

import (
	"context"
	"slices"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

const testTimeout = 2 * time.Second

func collect(t *testing.T, in <-chan int) []int {
	t.Helper()
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
			t.Fatal("timed out waiting for values and output closure")
			return nil
		}
	}
}

func wait(t *testing.T, done <-chan struct{}) {
	t.Helper()
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		t.Fatal("stages did not finish")
	}
}

// Cleanup отменяет и ожидает все стадии даже при раннем падении проверки.
func start(t *testing.T, numbers []int) (<-chan int, <-chan struct{}, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	out, done := pipeline(ctx, numbers)
	t.Cleanup(func() { cancel(); wait(t, done) })
	return out, done, cancel
}

func TestPipeline(t *testing.T) {
	for _, tt := range []struct {
		name          string
		numbers, want []int
	}{
		{name: "nil input"},
		{name: "empty input", numbers: []int{}},
		{name: "no matching values", numbers: []int{1, 3, 5}},
		{name: "single match", numbers: []int{2}, want: []int{4}},
		{name: "zero", numbers: []int{0}, want: []int{0}},
		{name: "order negatives duplicates", numbers: []int{-4, 1, 2, 0, -2, 3, 2}, want: []int{16, 4, 0, 4, 4}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, done, _ := start(t, tt.numbers)
			if got := collect(t, out); !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			wait(t, done)
		})
	}
}

func TestStageComposition(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	source := generate(ctx, &wg, []int{-3, 0, 2})
	out := square(ctx, &wg, square(ctx, &wg, source))
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	t.Cleanup(func() { cancel(); wait(t, done) })
	if got, want := collect(t, out), []int{81, 0, 16}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	wait(t, done)
}

func TestPipelineEarlyExit(t *testing.T) {
	out, done, cancel := start(t, []int{2, 4, 6, 8, 10})
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	select {
	case value, ok := <-out:
		if !ok || value != 4 {
			t.Fatalf("first result = (%d, %t), want (4, true)", value, ok)
		}
	case <-timer.C:
		t.Fatal("first result did not arrive")
	}
	cancel()
	// Не дочитываем выход, иначе можно скрыть заблокированную отправку.
	wait(t, done)
	if got := collect(t, out); len(got) != 0 {
		t.Fatalf("data remained after cancellation: %v", got)
	}
}

func TestPipelineAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out, done := pipeline(ctx, []int{2, 4, 6})
	wait(t, done)
	if got := collect(t, out); len(got) != 0 {
		t.Fatalf("cancelled source generated %v", got)
	}
}

func TestPipelineCancelWithUnreadOutput(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		out, done, cancel := start(t, []int{2, 4, 6, 8})
		// Все стадии продвинулись до устойчивой блокировки. Выход никто не читает.
		synctest.Wait()
		cancel()
		wait(t, done)
		if got := collect(t, out); len(got) != 0 {
			t.Fatalf("unexpected queued output: %v", got)
		}
	})
}

func TestStagesCancelWhileReadingNilInput(t *testing.T) {
	for _, tt := range []struct {
		name  string
		stage func(context.Context, *sync.WaitGroup, <-chan int) <-chan int
	}{
		{name: "square", stage: square},
		{name: "filter", stage: onlyEven},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				var wg sync.WaitGroup
				out := tt.stage(ctx, &wg, nil)
				done := make(chan struct{})
				go func() { wg.Wait(); close(done) }()
				synctest.Wait()
				cancel()
				wait(t, done)
				if got := collect(t, out); len(got) != 0 {
					t.Fatalf("unexpected output: %v", got)
				}
			})
		})
	}
}

func TestDeadlineStopsIdleStage(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		var wg sync.WaitGroup
		out := square(ctx, &wg, nil)
		if got := collect(t, out); len(got) != 0 {
			t.Fatalf("unexpected values: %v", got)
		}
		done := make(chan struct{})
		go func() { wg.Wait(); close(done) }()
		wait(t, done)
		if ctx.Err() != context.DeadlineExceeded {
			t.Fatalf("got %v, want deadline exceeded", ctx.Err())
		}
	})
}
