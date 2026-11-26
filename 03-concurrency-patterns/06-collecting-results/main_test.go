package main

import (
	"context"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

func drain(t *testing.T, results <-chan Result) []Result {
	t.Helper()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	var got []Result
	for {
		select {
		case result, ok := <-results:
			if !ok {
				return got
			}
			got = append(got, result)
		case <-timer.C:
			t.Fatal("timed out waiting for results and closure")
			return nil
		}
	}
}

func TestRunTasksDeliversOneResultPerIndex(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inputs := []int{3, 0, -4, 3}
	got := drain(t, runTasks(ctx, inputs, doublePositive))
	var ids []int
	for _, result := range got {
		ids = append(ids, result.Index)
	}
	slices.Sort(ids)
	if !slices.Equal(ids, []int{0, 1, 2, 3}) {
		t.Fatalf("unexpected result indices: %v", ids)
	}
	for _, result := range got {
		index := result.Index
		if index == 2 {
			if result.Err == nil {
				t.Fatal("expected an error for the negative input")
			}
		} else if result.Err != nil || result.Value != inputs[index]*2 {
			t.Fatalf("unexpected result: %+v", result)
		}
	}
}

func TestInputOrderDistinguishesMissingZeroAndError(t *testing.T) {
	results := make(chan Result, 3)
	_, jobErr := doublePositive(context.Background(), -1)
	results <- Result{Index: 3, Value: 6}
	results <- Result{Index: 2, Err: jobErr}
	results <- Result{Index: 0, Value: 0}
	close(results)
	got := inInputOrder(4, results)
	want := []Entry{{Received: true}, {}, {Received: true, Err: jobErr}, {Received: true, Value: 6}}
	if !slices.Equal(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestEmptyAndAlreadyCancelled(t *testing.T) {
	if got := drain(t, runTasks(context.Background(), nil, doublePositive)); len(got) != 0 {
		t.Fatalf("empty inputs produced %v", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := make(chan struct{}, 1)
	out := runTasks(ctx, []int{4}, func(context.Context, int) (int, error) {
		called <- struct{}{}
		return 8, nil
	})
	if got := drain(t, out); len(got) != 0 || len(called) != 0 {
		t.Fatalf("cancelled batch processed work: %v", got)
	}
}

func TestCancelWhilePublishingToUnreadOutput(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		results := runTasks(ctx, []int{2, 4}, doublePositive)
		synctest.Wait()
		cancel()
		synctest.Wait()
		// Проверяем закрытие только после выхода горутин из ожидания отправки.
		select {
		case result, ok := <-results:
			if ok {
				t.Fatalf("unexpected pending result: %+v", result)
			}
		default:
			t.Fatal("output stayed open after cancellation")
		}
	})
}
