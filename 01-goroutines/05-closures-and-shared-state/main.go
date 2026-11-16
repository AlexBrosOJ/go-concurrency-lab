package main

import (
	"fmt"
	"sync"
)

// Каждый запуск получает собственный параметр value. Запись в разные
// элементы заранее выделенного слайса не требует общего mutex.
func squares(values []int) []int {
	results := make([]int, len(values))
	var wg sync.WaitGroup
	for index, value := range values {
		wg.Add(1)
		go func(index, value int) {
			defer wg.Done()
			results[index] = value * value
		}(index, value)
	}
	// Wait нужен не только для ожидания: после него записи работников
	// видны читающей горутине. До Wait читать results нельзя.
	wg.Wait()
	return results
}

// Общая переменная остаётся общей даже при передаче её адреса параметром.
// Mutex защищает всю операцию read-modify-write, а не только присваивание.
func sharedCounter(workers, increments int) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	counter := 0
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < increments; i++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return counter
}

func main() {
	fmt.Println("=== 1. Closure captures a variable, not a snapshot ===")
	value := 10
	read := func() int { return value }
	value = 20
	fmt.Println("Closure sees:", read()) // 20

	fmt.Println("\n=== 2. Explicit snapshot ===")
	snapshot := value
	readSnapshot := func() int { return snapshot }
	value = 30
	fmt.Println("Original:", value, "snapshot:", readSnapshot()) // 30, 20

	fmt.Println("\n=== 3. Loop variables declared with := (Go 1.22+) ===")
	var readers []func() int
	for i := 0; i < 3; i++ {
		readers = append(readers, func() int { return i })
	}
	for _, reader := range readers {
		fmt.Println(reader()) // 0, 1, 2
	}

	fmt.Println("\n=== 4. Assignment to an existing loop variable ===")
	readers = nil
	var i int
	for i = 0; i < 3; i++ {
		readers = append(readers, func() int { return i })
	}
	for _, reader := range readers {
		fmt.Println(reader()) // 3, 3, 3: одна переменная на весь цикл
	}

	fmt.Println("\n=== 5. Goroutine arguments and independent result slots ===")
	fmt.Println("Squares:", squares([]int{2, 3, 4})) // [4 9 16]

	fmt.Println("\n=== 6. A copied slice still shares its backing array ===")
	original := []int{10, 20}
	alias := original
	alias[0] = 99
	fmt.Println("Original after alias write:", original)
	independent := append([]int(nil), original...)
	independent[0] = -1
	fmt.Println("Original:", original, "independent:", independent)

	fmt.Println("\n=== 7. Channel establishes an execution order ===")
	ready := make(chan struct{})
	finished := make(chan struct{})
	message := "before"
	go func() {
		defer close(finished)
		<-ready
		fmt.Println("Worker sees:", message)
	}()
	message = "after"
	close(ready) // запись message происходит до разблокировки получателя
	<-finished   // main не завершится до окончания работника

	fmt.Println("\n=== 8. Shared state protected by a mutex ===")
	fmt.Println("Counter:", sharedCounter(4, 1000)) // 4000
	fmt.Println("WaitGroup waits; mutex protects concurrent updates.")
}
