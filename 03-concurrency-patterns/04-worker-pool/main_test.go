package main

import (
	"context"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

func receive[T any](t *testing.T, ch <-chan T) (T, bool) {
	t.Helper()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case value, ok := <-ch:
		return value, ok
	case <-timer.C:
		t.Fatal("timed out receiving from channel")
		var zero T
		return zero, false
	}
}

func collect(t *testing.T, out <-chan Result) []Result {
	t.Helper()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	var results []Result
	for {
		select {
		case result, ok := <-out:
			if !ok {
				return results
			}
			results = append(results, result)
		case <-timer.C:
			t.Fatal("timed out waiting for pool output to close")
			return nil
		}
	}
}

func TestPoolResults(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	batch := []Job{{1, "one"}, {2, "  "}, {3, "два слова"}, {4, "one"}}
	out, err := startPool(ctx, batch, 2, 1, countWords)
	if err != nil {
		t.Fatal(err)
	}
	var ids []int
	for _, result := range collect(t, out) {
		ids = append(ids, result.ID)
		if result.ID == 2 {
			if result.Err == nil {
				t.Fatal("expected an error for blank text")
			}
			continue
		}
		want := 1
		if result.ID == 3 {
			want = 2
		}
		if result.Err != nil || result.Words != want {
			t.Fatalf("got %+v, want %d words and no error", result, want)
		}
	}
	slices.Sort(ids)
	if !slices.Equal(ids, []int{1, 2, 3, 4}) {
		t.Fatalf("unexpected result IDs: %v", ids)
	}
}

func TestPoolEmptyBatch(t *testing.T) {
	out, err := startPool(context.Background(), nil, 3, 0, countWords)
	if err != nil {
		t.Fatal(err)
	}
	if results := collect(t, out); len(results) != 0 {
		t.Fatalf("unexpected results: %v", results)
	}
}

func TestPoolRejectsInvalidConfiguration(t *testing.T) {
	for _, config := range [][2]int{{0, 0}, {-1, 1}, {2, -1}} {
		out, err := startPool(context.Background(), nil, config[0], config[1], countWords)
		if err == nil || out != nil {
			t.Fatalf("configuration %v should fail before starting a pool", config)
		}
	}
}

func TestPoolLimitsActiveJobs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		batch := make([]Job, 6)
		started := make(chan struct{}, len(batch))
		release := make(chan struct{})
		out, err := startPool(ctx, batch, 2, 3, func(ctx context.Context, _ Job) (int, error) {
			started <- struct{}{}
			select {
			case <-release:
				return 1, nil
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		if got := len(started); got != 2 {
			t.Fatalf("started %d jobs with all active processors held, want 2", got)
		}
		close(release)
		if got := len(collect(t, out)); got != len(batch) {
			t.Fatalf("got %d results, want %d", got, len(batch))
		}
	})
}

func TestPoolCancellationWithoutOutputReader(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		processed := make(chan struct{}, 1)
		out, err := startPool(ctx, []Job{{ID: 1}, {ID: 2}, {ID: 3}}, 1, 0, func(context.Context, Job) (int, error) {
			processed <- struct{}{}
			return 1, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		receive(t, processed)
		synctest.Wait()
		cancel()
		synctest.Wait()
		// Проверяем выход после отмены, чтобы чтение не разблокировало отправку раньше.
		select {
		case result, ok := <-out:
			if ok {
				t.Fatalf("pool still had a blocked result: %+v", result)
			}
		default:
			t.Fatal("pool failed to close while nobody was reading")
		}
	})
}

func TestPoolPassesCancellationToProcessor(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		started := make(chan struct{})
		stopped := make(chan error, 1)
		out, err := startPool(ctx, []Job{{ID: 1}}, 1, 0, func(ctx context.Context, _ Job) (int, error) {
			close(started)
			<-ctx.Done()
			stopped <- ctx.Err()
			return 0, ctx.Err()
		})
		if err != nil {
			t.Fatal(err)
		}
		receive(t, started)
		cancel()
		if got, ok := receive(t, stopped); !ok || got != context.Canceled {
			t.Fatalf("processor reported %v, want context.Canceled", got)
		}
		synctest.Wait()
		if result, ok := receive(t, out); ok {
			t.Fatalf("unexpected result after pool stopped: %+v", result)
		}
	})
}
