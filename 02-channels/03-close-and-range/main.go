package main

import (
	"fmt"
	"sync"
)

// stream создаёт конечный поток. Единственный производитель владеет закрытием.
// Потребитель должен дочитать канал: отмена в этом примере не реализована.
func stream(values []int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, value := range values {
			out <- value
		}
	}()
	return out
}

// sum читает поток неизвестной для получателя длины до закрытия канала.
func sum(in <-chan int) int {
	total := 0
	for value := range in {
		total += value
	}
	return total
}

// sharedStream запускает по производителю на группу значений.
// Закрывает общий выход только координатор после завершения всех отправок.
func sharedStream(groups ...[]int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(groups))
	for _, values := range groups {
		go func() {
			defer wg.Done()
			for _, value := range values {
				out <- value
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func finiteStreamExample() {
	for value := range stream([]int{10, 20, 30}) {
		fmt.Println("Received:", value)
	}
	fmt.Println("Range finished: stream closed and drained")
}

func commaOKExample() {
	values := make(chan int, 2)
	values <- 0
	values <- 42
	close(values)
	for i := 0; i < 4; i++ {
		value, ok := <-values
		fmt.Printf("Read %d: value=%d ok=%t\n", i+1, value, ok)
	}
}

func emptyStreamExample() {
	count := 0
	for range stream(nil) {
		count++
	}
	fmt.Println("Empty stream iterations:", count)
}

func sumExample() {
	fmt.Println("Sum:", sum(stream([]int{1, 2, 3, 4})))
}

func multipleProducersExample() {
	// Отправки двух производителей могут перемешиваться.
	count, total := 0, 0
	for value := range sharedStream([]int{1, 3, 5}, []int{2, 4, 6}) {
		count++
		total += value
	}
	fmt.Printf("All producers finished: count=%d sum=%d\n", count, total)
}

func multipleConsumersExample() {
	values := stream([]int{1, 2, 3, 4, 5, 6})
	results := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func() {
			results <- sum(values)
		}()
	}
	// По одному итоговому сообщению от каждого обработчика.
	// Сам results закрывать не нужно: число получений заранее известно.
	total := <-results + <-results
	fmt.Println("Two consumers, combined sum:", total)
}

func broadcastExample() {
	done := make(chan struct{})
	acknowledged := make(chan struct{}, 3)
	for i := 0; i < 3; i++ {
		go func() {
			<-done
			acknowledged <- struct{}{}
		}()
	}
	close(done)
	for i := 0; i < 3; i++ {
		<-acknowledged
	}
	fmt.Println("Closed signal observed by all 3 listeners")
}

// recover нужен только для демонстрации ошибок, чтобы урок мог идти дальше.
// В рабочем протоколе закрытия паники следует предотвращать.
func showPanic(label string, action func()) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%s: %v\n", label, err)
		}
	}()
	action()
}

func invalidOperationsExample() {
	closed := make(chan int, 1)
	close(closed)
	showPanic("Send after close", func() { closed <- 1 })
	showPanic("Close twice", func() { close(closed) })
	var nilChannel chan int
	showPanic("Close nil", func() { close(nilChannel) })
}

func main() {
	fmt.Println("=== 1. A finite stream and range ===")
	finiteStreamExample()
	fmt.Println("\n=== 2. Buffered values and comma-ok ===")
	commaOKExample()
	fmt.Println("\n=== 3. An empty stream ===")
	emptyStreamExample()
	fmt.Println("\n=== 4. An aggregate over a stream ===")
	sumExample()
	fmt.Println("\n=== 5. Multiple producers, one closer ===")
	multipleProducersExample()
	fmt.Println("\n=== 6. Multiple consumers ===")
	multipleConsumersExample()
	fmt.Println("\n=== 7. Close as a broadcast signal ===")
	broadcastExample()
	fmt.Println("\n=== 8. Invalid operations (recovered for demonstration) ===")
	invalidOperationsExample()
}
