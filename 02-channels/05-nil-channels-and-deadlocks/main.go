package main

import "fmt"

// drainTwo объединяет два потока до их закрытия. nil означает отключённый
// вход. Производители обязаны закрыть ненулевые входы; отмены здесь нет.
func drainTwo(first, second <-chan int) []int {
	var results []int
	for first != nil || second != nil {
		select {
		case value, ok := <-first:
			if !ok {
				first = nil
				continue
			}
			results = append(results, value)
		case value, ok := <-second:
			if !ok {
				second = nil
				continue
			}
			results = append(results, value)
		}
	}
	return results
}

func main() {
	fmt.Println("=== 1. Zero-value channel ===")
	var absent chan int
	fmt.Println("Channel is nil:", absent == nil)
	fmt.Println("Length:", len(absent), "capacity:", cap(absent))

	fmt.Println("\n=== 2. A nil channel cannot send or receive ===")
	// Без default обе операции заблокировали бы эту горутину навсегда.
	select {
	case absent <- 42:
		fmt.Println("Unreachable send")
	case value := <-absent:
		fmt.Println("Unreachable receive:", value)
	default:
		fmt.Println("Neither nil-channel operation is ready")
	}

	fmt.Println("\n=== 3. nil and closed are different states ===")
	closed := make(chan int, 1)
	closed <- 7
	close(closed)
	value, ok := <-closed
	fmt.Println("Buffered value:", value, "ok:", ok) // 7, true
	value, ok = <-closed
	fmt.Println("After draining:", value, "ok:", ok) // 0, false
	// Закрытый пустой канал всегда готов к получению; nil — никогда.
	select {
	case <-absent:
		fmt.Println("Unreachable nil receive")
	case value, ok := <-closed:
		fmt.Println("Closed channel selected:", value, ok)
	}

	fmt.Println("\n=== 4. Disable an optional select branch ===")
	updates := make(chan int, 1)
	updates <- 10
	var optional <-chan int
	select {
	case <-optional:
		fmt.Println("Unreachable optional branch")
	case value := <-updates:
		fmt.Println("Active branch:", value)
	}

	fmt.Println("\n=== 5. Disable closed inputs while draining ===")
	first := make(chan int, 2)
	second := make(chan int, 2)
	first <- 1
	first <- 2
	second <- 10
	second <- 20
	close(first)
	close(second)
	// Порядок между каналами не определён, но внутри каждого сохраняется.
	fmt.Println("Merged:", drainTwo(first, second))
	fmt.Println("No inputs:", drainTwo(nil, nil))

	fmt.Println("\n=== 6. Unbuffered handoff needs another participant ===")
	handoff := make(chan string)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		handoff <- "message from sender"
	}()
	fmt.Println("Received:", <-handoff)
	<-finished

	fmt.Println("\n=== 7. A buffer has a finite capacity ===")
	queue := make(chan int, 1)
	queue <- 100
	select {
	case queue <- 200:
		fmt.Println("Unreachable: queue is full")
	default:
		fmt.Println("Send would block: queue is full")
	}
	fmt.Println("Free a slot:", <-queue)
	queue <- 200
	fmt.Println("Next value:", <-queue)

	fmt.Println("\n=== 8. Deadlock examples are documented, not executed ===")
	fmt.Println("No sleeps, leaked goroutines, or intentional deadlocks in this demo.")
}
