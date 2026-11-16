package main

import (
	"errors"
	"fmt"
	"time"
)

var errTimeout = errors.New("operation timed out")

// receiveWithin различает данные, закрытие входа и истечение времени.
// Таймаут ограничивает это получение, но не отменяет производителя.
func receiveWithin(in <-chan int, timeout time.Duration) (int, bool, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case value, ok := <-in:
		return value, ok, nil
	case <-timer.C:
		return 0, false, errTimeout
	}
}

// sendWithin ограничивает ожидание отправки. out не должен быть закрыт.
func sendWithin(out chan<- int, value int, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case out <- value:
		return nil
	case <-timer.C:
		return errTimeout
	}
}

// mergeUntilDone объединяет два потока с возможностью остановить пересылку.
// Входы закрывают их производители; выход закрывает только эта горутина.
// Отмена пересылки не отменяет внешних производителей автоматически.
func mergeUntilDone(done <-chan struct{}, first, second <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for first != nil || second != nil {
			var value int
			var ok bool
			select {
			case <-done:
				return
			case value, ok = <-first:
				if !ok {
					first = nil // Отключаем закрытый вход в следующих select.
					continue
				}
			case value, ok = <-second:
				if !ok {
					second = nil
					continue
				}
			}
			// Отмена нужна и здесь: потребитель мог перестать читать out.
			select {
			case out <- value:
			case <-done:
				return
			}
		}
	}()
	return out
}

func readyCasesExample() {
	first := make(chan string, 1)
	second := make(chan string, 1)
	first <- "first"
	second <- "second"
	select {
	case value := <-first:
		fmt.Println("Selected:", value)
	case value := <-second:
		fmt.Println("Selected:", value)
	}
	fmt.Println("Only one case runs; neither source has priority")
}

func defaultExample() {
	values := make(chan int)
	select {
	case value := <-values:
		fmt.Println("Received:", value)
	default:
		fmt.Println("Nothing ready: continue without waiting")
	}
}

func selectSendExample() {
	values := make(chan int, 1)
	select {
	case values <- 42:
		fmt.Println("Queued:", <-values)
	default:
		fmt.Println("No free slot")
	}
}

func receiveAndCloseExample() {
	values := make(chan int, 1)
	values <- 0
	close(values)
	for i := 0; i < 2; i++ {
		value, ok, err := receiveWithin(values, time.Second)
		fmt.Printf("Receive: value=%d ok=%t err=%v\n", value, ok, err)
	}
}

func receiveTimeoutExample() {
	// nil отключает ветку получения: готовым станет только таймер.
	value, ok, err := receiveWithin(nil, 10*time.Millisecond)
	fmt.Printf("No source: value=%d ok=%t err=%v\n", value, ok, err)
}

func sendTimeoutExample() {
	values := make(chan int, 1)
	values <- 10
	err := sendWithin(values, 20, 10*time.Millisecond)
	fmt.Println("Full buffer:", err)
	fmt.Println("Original value:", <-values)
}

func timeAfterExample() {
	var values <-chan int
	select {
	case value := <-values:
		fmt.Println("Received:", value)
	case <-time.After(10 * time.Millisecond):
		fmt.Println("time.After: waiting ended")
	}
}

func disableClosedInputExample() {
	first := make(chan int, 2)
	first <- 1
	first <- 3
	close(first)
	second := make(chan int, 2)
	second <- 2
	second <- 4
	close(second)
	count, total := 0, 0
	for value := range mergeUntilDone(nil, first, second) {
		count++
		total += value
	}
	fmt.Printf("Both sources drained: count=%d sum=%d\n", count, total)
}

func cancellationExample() {
	done := make(chan struct{})
	first := make(chan int, 1)
	first <- 7
	second := make(chan int) // Открытый пустой вход тоже может ждать бесконечно.
	out := mergeUntilDone(done, first, second)
	// До отмены выход никто не читает. Пересылка могла ждать чтение или отправку.
	close(done)
	for range out {
		// Готовые операции могут состязаться с отменой: дочитываем выход.
	}
	fmt.Println("Cancelled merge finished without closing its inputs")
}

func main() {
	fmt.Println("=== 1. Two ready cases ===")
	readyCasesExample()
	fmt.Println("\n=== 2. Nonblocking receive with default ===")
	defaultExample()
	fmt.Println("\n=== 3. A send case ===")
	selectSendExample()
	fmt.Println("\n=== 4. Value and channel closure ===")
	receiveAndCloseExample()
	fmt.Println("\n=== 5. Receive timeout ===")
	receiveTimeoutExample()
	fmt.Println("\n=== 6. Send timeout ===")
	sendTimeoutExample()
	fmt.Println("\n=== 7. One-shot time.After ===")
	timeAfterExample()
	fmt.Println("\n=== 8. Disable closed inputs with nil ===")
	disableClosedInputExample()
	fmt.Println("\n=== 9. Cancel reads and writes ===")
	cancellationExample()
}
