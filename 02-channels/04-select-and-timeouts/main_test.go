package main

import (
	"errors"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

const testTimeout = 2 * time.Second

// Страхуем даже сами функции с таймаутом от ошибки, которая заставит их зависнуть.
func bounded[T any](t *testing.T, action func() T) T {
	t.Helper()
	result := make(chan T, 1)
	go func() { result <- action() }()
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	select {
	case value := <-result:
		return value
	case <-timer.C:
		t.Fatal("operation exceeded test watchdog")
		var zero T
		return zero
	}
}

type received struct {
	value int
	ok    bool
	err   error
}

func TestReceiveWithin(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		closed bool
		want   received
	}{
		{name: "positive", values: []int{42}, want: received{value: 42, ok: true}},
		{name: "zero is data", values: []int{0}, want: received{ok: true}},
		{name: "negative", values: []int{-3}, want: received{value: -3, ok: true}},
		{name: "closed empty", closed: true},
		{name: "closed with buffered data", values: []int{7}, closed: true, want: received{value: 7, ok: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan int, len(tt.values))
			for _, value := range tt.values {
				ch <- value
			}
			if tt.closed {
				close(ch)
			}
			got := bounded(t, func() received {
				value, ok, err := receiveWithin(ch, time.Hour)
				return received{value, ok, err}
			})
			if got.value != tt.want.value || got.ok != tt.want.ok || got.err != nil {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestReceiveWithinTimeout(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ch      <-chan int
		timeout time.Duration
	}{
		{name: "nil input", timeout: 5 * time.Millisecond},
		{name: "open empty input", ch: make(chan int), timeout: 5 * time.Millisecond},
		{name: "zero duration", timeout: 0},
		{name: "negative duration", timeout: -time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := bounded(t, func() received {
				value, ok, err := receiveWithin(tc.ch, tc.timeout)
				return received{value, ok, err}
			})
			if !errors.Is(got.err, errTimeout) || got.ok || got.value != 0 {
				t.Fatalf("got %+v, want timeout without a value", got)
			}
		})
	}
}

func TestSendWithin(t *testing.T) {
	ch := make(chan int, 1)
	if err := bounded(t, func() error { return sendWithin(ch, 0, time.Hour) }); err != nil {
		t.Fatal(err)
	}
	select {
	case value := <-ch:
		if value != 0 {
			t.Fatalf("received %d, want 0", value)
		}
	default:
		t.Fatal("successful send did not queue a value")
	}
}

func TestSendWithinTimeout(t *testing.T) {
	full := make(chan int, 1)
	full <- 10
	for _, tc := range []struct {
		name string
		ch   chan<- int
	}{
		{name: "full buffer", ch: full},
		{name: "nil output"},
		{name: "unbuffered without receiver", ch: make(chan int)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := bounded(t, func() error { return sendWithin(tc.ch, 20, 5*time.Millisecond) }); !errors.Is(err, errTimeout) {
				t.Fatalf("got %v, want timeout", err)
			}
		})
	}
	if got := <-full; got != 10 {
		t.Fatalf("queued value changed to %d", got)
	}
	select {
	case value := <-full:
		t.Fatalf("timed-out send queued %d", value)
	default:
	}
}

func buffered(values []int) <-chan int {
	ch := make(chan int, len(values))
	for _, value := range values {
		ch <- value
	}
	close(ch)
	return ch
}

func collect(t *testing.T, in <-chan int) []int {
	t.Helper()
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	var got []int
	for {
		select {
		case value, ok := <-in:
			if !ok {
				return got
			}
			got = append(got, value)
		case <-timer.C:
			t.Fatal("output did not close")
			return nil
		}
	}
}

func TestMergeUntilDone(t *testing.T) {
	tests := []struct {
		name          string
		first, second <-chan int
		want          []int
	}{
		{name: "both nil"},
		{name: "both empty closed", first: buffered(nil), second: buffered(nil)},
		{name: "nil and data", second: buffered([]int{3, 1}), want: []int{1, 3}},
		{name: "closed and data", first: buffered(nil), second: buffered([]int{4}), want: []int{4}},
		{name: "duplicates and negatives", first: buffered([]int{0, -1, 2}), second: buffered([]int{2, -1}), want: []int{-1, -1, 0, 2, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})
			defer close(done) // Освобождаем пересылку даже при падении проверки.
			got := collect(t, mergeUntilDone(done, tt.first, tt.second))
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeCancellationWhileReading(t *testing.T) {
	done := make(chan struct{})
	first, second := make(chan int), make(chan int)
	out := mergeUntilDone(done, first, second)
	close(done)
	if got := collect(t, out); len(got) != 0 {
		t.Fatalf("unexpected data: %v", got)
	}
	// Пересылка не должна закрывать чужие входы.
	for _, ch := range []chan int{first, second} {
		select {
		case _, ok := <-ch:
			if !ok {
				t.Fatal("merge closed an input")
			}
		default:
		}
		close(ch)
	}
}

func TestMergeCancellationWhileSending(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		done := make(chan struct{})
		first := make(chan int, 1)
		first <- 42
		close(first)
		out := mergeUntilDone(done, first, nil)
		// Wait возвращается, когда пересылка устойчиво заблокирована на выходе.
		synctest.Wait()
		close(done)
		synctest.Wait()
		// До этой проверки получателя не было: отмена должна сама завершить отправку.
		select {
		case value, ok := <-out:
			if ok {
				t.Fatalf("pending value %d was not abandoned on cancellation", value)
			}
		default:
			t.Fatal("output stayed open after cancellation")
		}
	})
}
