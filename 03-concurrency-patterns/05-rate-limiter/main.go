package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// rateLimiter ограничивает количество одновременных операций значением limit.
// Реализация через буферизованный канал-семафор: перед вызовом горoutine
// "забирает разрешение" из канала, после вызова — возвращает его обратно.
func rateLimiter[T any](jobs []T, limit int, process func(T)) {
	tokens := make(chan struct{}, limit)

	// Наполняем семафор разрешениями.
	for i := 0; i < limit; i++ {
		tokens <- struct{}{}
	}

	var wg sync.WaitGroup

	for _, job := range jobs {
		wg.Add(1)

		go func(job T) {
			defer wg.Done()

			<-tokens       // ждём свободное разрешение
			defer func() { // всегда возвращаем разрешение, даже при панике
				tokens <- struct{}{}
			}()

			process(job)
		}(job)
	}

	wg.Wait()
}

func main() {
	const limit = 3

	var inFlight atomic.Int64
	var maxInFlight atomic.Int64

	jobs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	start := time.Now()

	rateLimiter(jobs, limit, func(id int) {
		current := inFlight.Add(1)

		// Запоминаем максимальное число одновременно работающих горутин.
		for {
			previous := maxInFlight.Load()
			if current <= previous || maxInFlight.CompareAndSwap(previous, current) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond)
		inFlight.Add(-1)

		fmt.Printf("задача %d выполнена\n", id)
	})

	fmt.Printf("выполнено %d задач за %v\n", len(jobs), time.Since(start).Round(time.Millisecond))
	fmt.Printf("максимум одновременно выполняемых задач: %d (лимит: %d)\n",
		maxInFlight.Load(), limit)
}
