package main

import (
	"slices"
	"strconv"
	"testing"
	"time"
)

const testTimeout = 2 * time.Second

func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	select {
	case value := <-ch:
		return value
	case <-timer.C:
		t.Fatal("timed out waiting for channel receive")
		var zero T
		return zero
	}
}

// Если проверка завершится раньше чтения, cleanup примет оставшуюся отправку.
// Ожидание cleanup тоже ограничено, чтобы ошибка не повесила весь пакет.
func startSender(t *testing.T, ch chan int, value int) <-chan struct{} {
	t.Helper()
	return startProducer(t, ch, func() { sendValue(ch, value) })
}

func startProducer(t *testing.T, ch chan int, send func()) <-chan struct{} {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		send()
	}()
	t.Cleanup(func() {
		timer := time.NewTimer(testTimeout)
		defer timer.Stop()
		for {
			select {
			case <-done:
				return
			case <-ch:
			case <-timer.C:
				t.Error("sender did not finish during cleanup")
				return
			}
		}
	})
	return done
}

func TestSendValue(t *testing.T) {
	for _, want := range []int{42, 0, -7} {
		t.Run(strconv.Itoa(want), func(t *testing.T) {
			ch := make(chan int)
			done := startSender(t, ch, want)
			if got := receive(t, ch); got != want {
				t.Fatalf("received %d, want %d", got, want)
			}
			receive(t, done)
			// После завершения отправителя проверяем отсутствие второй отправки.
			select {
			case value := <-ch:
				t.Fatalf("unexpected extra value: %d", value)
			default:
			}
		})
	}
}

func TestUnbufferedSendNeedsReceiver(t *testing.T) {
	ch := make(chan int)
	if cap(ch) != 0 {
		t.Fatal("expected an unbuffered channel")
	}
	// Пока получателя нет, отправка не готова: выполняется default.
	// Не нужно угадывать задержку, за которую планировщик запустит горутину.
	select {
	case ch <- 42:
		t.Fatal("send succeeded without a receiver")
	default:
	}
	done := startSender(t, ch, 42)
	select {
	case <-done:
		t.Fatal("sendValue returned before a receive")
	default:
	}
	if got := receive(t, ch); got != 42 {
		t.Fatalf("received %d, want 42", got)
	}
	receive(t, done)
}

func TestUnbufferedReceiveNeedsSender(t *testing.T) {
	ch := make(chan int)
	// Пустой открытый канал не возвращает нулевое значение без отправителя.
	select {
	case value := <-ch:
		t.Fatalf("receive succeeded without a sender: %d", value)
	default:
	}
	done := startSender(t, ch, 0)
	if got := receive(t, ch); got != 0 {
		t.Fatalf("received %d, want 0", got)
	}
	receive(t, done)
}

func TestSendValues(t *testing.T) {
	tests := []struct {
		name   string
		values []int
	}{
		{name: "nil slice"},
		{name: "empty slice", values: []int{}},
		{name: "single value", values: []int{42}},
		{name: "order and duplicates", values: []int{3, 0, -1, 3, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan int)
			done := startProducer(t, ch, func() { sendValues(ch, tt.values) })
			var got []int
			for range tt.values {
				got = append(got, receive(t, ch))
			}
			receive(t, done)
			if !slices.Equal(got, tt.values) {
				t.Fatalf("received %v, want %v in the same order", got, tt.values)
			}
			select {
			case value := <-ch:
				t.Fatalf("unexpected extra value: %d", value)
			default:
			}
		})
	}
}

func TestSquareOnce(t *testing.T) {
	for _, value := range []int{7, 0, -3} {
		t.Run(strconv.Itoa(value), func(t *testing.T) {
			requests := make(chan int)
			responses := make(chan int)
			done := make(chan struct{})
			go func() {
				defer close(done)
				squareOnce(requests, responses)
			}()
			t.Cleanup(func() {
				timer := time.NewTimer(testTimeout)
				defer timer.Stop()
				for {
					select {
					case <-done:
						return
					case requests <- value:
					case <-responses:
					case <-timer.C:
						t.Error("request handler did not finish during cleanup")
						return
					}
				}
			})
			timer := time.NewTimer(testTimeout)
			defer timer.Stop()
			select {
			case requests <- value:
			case <-timer.C:
				t.Fatal("timed out sending request")
			}
			// Без получения ответа обработчик не должен завершиться.
			select {
			case <-done:
				t.Fatal("handler finished before its response was received")
			default:
			}
			if got, want := receive(t, responses), value*value; got != want {
				t.Fatalf("response = %d, want %d", got, want)
			}
			receive(t, done)
		})
	}
}
