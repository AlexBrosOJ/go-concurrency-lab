package main

import (
	"slices"
	"strconv"
	"testing"
	"time"
)

const testTimeout = 2 * time.Second

func send[T any](t *testing.T, ch chan<- T, value T) {
	t.Helper()
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	select {
	case ch <- value:
	case <-timer.C:
		t.Fatal("timed out sending value")
	}
}

func receive[T any](t *testing.T, ch <-chan T) (T, bool) {
	t.Helper()
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	select {
	case value, ok := <-ch:
		return value, ok
	case <-timer.C:
		t.Fatal("timed out receiving value")
		var zero T
		return zero, false
	}
}

func TestBufferCapacityAndFIFO(t *testing.T) {
	for _, capacity := range []int{1, 2, 5} {
		t.Run(strconv.Itoa(capacity), func(t *testing.T) {
			ch := make(chan int, capacity)
			// Отправки выполняются без отдельного получателя.
			for i := 0; i < capacity; i++ {
				send(t, ch, i)
			}
			if len(ch) != capacity || cap(ch) != capacity {
				t.Fatalf("len=%d cap=%d, want both %d", len(ch), cap(ch), capacity)
			}
			select {
			case ch <- 99:
				t.Fatal("send succeeded into a full buffer without a receiver")
			default:
			}
			for i := 0; i < capacity; i++ {
				if got, ok := receive(t, ch); !ok || got != i {
					t.Fatalf("received (%d, %t), want (%d, true)", got, ok, i)
				}
			}
			select {
			case <-ch:
				t.Fatal("receive succeeded from an empty open channel")
			default:
			}
		})
	}
}

func TestReceiveFreesBufferSlot(t *testing.T) {
	ch := make(chan int, 2)
	send(t, ch, 10)
	send(t, ch, 20)
	if got, _ := receive(t, ch); got != 10 {
		t.Fatalf("received %d, want 10", got)
	}
	send(t, ch, 30)
	for _, want := range []int{20, 30} {
		if got, _ := receive(t, ch); got != want {
			t.Fatalf("received %d, want %d", got, want)
		}
	}
}

func TestSendValues(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		values   []int
	}{
		{name: "nil input", capacity: 1},
		{name: "empty input", capacity: 2, values: []int{}},
		{name: "one slot", capacity: 1, values: []int{3, 0, -1, 3}},
		{name: "smaller buffer than input", capacity: 2, values: []int{5, 4, 3, 2, 1}},
		{name: "input fits buffer", capacity: 5, values: []int{0, -7, 0}},
		{name: "zero capacity is unbuffered", capacity: 0, values: []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan int, tt.capacity)
			done := make(chan struct{})
			go func() {
				defer close(done)
				sendValues(ch, tt.values)
			}()
			t.Cleanup(func() {
				timer := time.NewTimer(testTimeout)
				defer timer.Stop()
				for {
					select {
					case <-done:
						return
					case <-ch: // Освобождаем отправителя при раннем завершении теста.
					case <-timer.C:
						t.Error("producer did not finish during cleanup")
						return
					}
				}
			})
			var got []int
			for range tt.values {
				value, ok := receive(t, ch)
				if !ok {
					t.Fatal("channel closed before all values were received")
				}
				got = append(got, value)
			}
			receive(t, done)
			if !slices.Equal(got, tt.values) {
				t.Fatalf("received %v, want %v in order", got, tt.values)
			}
			select {
			case <-ch:
				t.Fatal("unexpected extra value or channel closure")
			default:
			}
		})
	}
}

func TestClosePreservesBufferedValues(t *testing.T) {
	ch := make(chan int, 3)
	for _, value := range []int{0, -1, 42} {
		send(t, ch, value)
	}
	close(ch)
	for _, want := range []int{0, -1, 42} {
		if got, ok := receive(t, ch); !ok || got != want {
			t.Fatalf("received (%d, %t), want (%d, true)", got, ok, want)
		}
	}
	// Закрытый опустошённый канал остаётся готовым к чтению.
	for i := 0; i < 2; i++ {
		if got, ok := receive(t, ch); ok || got != 0 {
			t.Fatalf("received (%d, %t), want (0, false)", got, ok)
		}
	}
}

func TestCloseEmptyBuffer(t *testing.T) {
	ch := make(chan int, 2)
	close(ch)
	if got, ok := receive(t, ch); ok || got != 0 {
		t.Fatalf("received (%d, %t), want (0, false)", got, ok)
	}
}
