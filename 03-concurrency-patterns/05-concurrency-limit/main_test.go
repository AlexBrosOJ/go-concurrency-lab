package main

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func await[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case value, ok := <-ch:
		if !ok {
			t.Fatal("channel closed without a value")
		}
		return value
	case <-timer.C:
		t.Fatal("timed out waiting for a value")
		var zero T
		return zero
	}
}

func TestRunLimitedBoundsActiveCallbacks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		inputs := []int{4, 1, 3, 2, 5}
		started := make(chan int, len(inputs))
		release := make(chan struct{})
		openRelease := sync.OnceFunc(func() { close(release) })
		defer openRelease()
		done := make(chan error, 1)
		go func() {
			done <- runLimited(ctx, inputs, 2, func(_ context.Context, input int) {
				started <- input
				<-release
			})
		}()

		synctest.Wait()
		if got := len(started); got != 2 {
			t.Fatalf("started %d callbacks before release, want 2", got)
		}
		openRelease()
		if err := await(t, done); err != nil {
			t.Fatal(err)
		}

		var got []int
		for range len(started) {
			got = append(got, <-started)
		}
		slices.Sort(got)
		if !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
			t.Fatalf("processed %v, want each input once", got)
		}
	})
}

func TestRunLimitedCancellationWaitsForStartedCallback(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		release := make(chan struct{})
		openRelease := sync.OnceFunc(func() { close(release) })
		defer openRelease()
		var started atomic.Int32
		done := make(chan error, 1)

		go func() {
			done <- runLimited(ctx, []int{1, 2, 3}, 1, func(ctx context.Context, _ int) {
				started.Add(1)
				<-ctx.Done()
				<-release // Имитируем завершение работы после получения сигнала отмены.
			})
		}()
		synctest.Wait()
		if got := started.Load(); got != 1 {
			t.Fatalf("started %d callbacks, want 1", got)
		}
		cancel()
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("returned before the active callback completed cleanup")
		default:
		}

		openRelease()
		if err := await(t, done); err != context.Canceled {
			t.Fatalf("got %v, want context.Canceled", err)
		}
		if got := started.Load(); got != 1 {
			t.Fatalf("started %d callbacks after cancellation, want 1", got)
		}
	})
}

func TestAcquireDeadlinePreservesSlotAccounting(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		slots := make(semaphore, 1)
		slots <- struct{}{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := slots.acquire(ctx); err != context.DeadlineExceeded {
			t.Fatalf("got %v, want context.DeadlineExceeded", err)
		}
		if got := len(slots); got != 1 {
			t.Fatalf("occupied slots = %d, want 1", got)
		}
		slots.release()
		if err := slots.acquire(context.Background()); err != nil {
			t.Fatal(err)
		}
		slots.release()
	})
}

func TestRunLimitedBoundaries(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tc := range []struct {
		name   string
		ctx    context.Context
		inputs []int
		limit  int
		want   error
	}{
		{name: "empty", ctx: context.Background(), limit: 2},
		{name: "zero limit", ctx: context.Background(), inputs: []int{1}, want: errInvalidLimit},
		{name: "negative limit", ctx: context.Background(), limit: -1, want: errInvalidLimit},
		{name: "already cancelled", ctx: cancelled, inputs: []int{1}, limit: 2, want: context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			done := make(chan error, 1)
			go func() {
				done <- runLimited(tc.ctx, tc.inputs, tc.limit, func(context.Context, int) {
					calls.Add(1)
				})
			}()
			if got := await(t, done); got != tc.want || calls.Load() != 0 {
				t.Fatalf("got %v and %d calls, want %v and 0 calls", got, calls.Load(), tc.want)
			}
		})
	}
}
