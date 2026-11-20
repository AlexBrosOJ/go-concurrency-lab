package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Каждый элемент в канале обозначает одно занятое разрешение.
type semaphore chan struct{}

func (s semaphore) acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case s <- struct{}{}:
		// Если разрешение получено одновременно с отменой, освобождаем место.
		if err := ctx.Err(); err != nil {
			s.release()
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s semaphore) release() {
	<-s
}

var errInvalidLimit = errors.New("limit must be positive")

// runLimited ждёт все уже запущенные задачи даже после отмены.
// Для досрочного выхода обработчик сам проверяет контекст.
func runLimited[T any](ctx context.Context, inputs []T, limit int, process func(context.Context, T)) error {
	if limit < 1 {
		return errInvalidLimit
	}

	slots := make(semaphore, limit)
	var wg sync.WaitGroup
	for _, input := range inputs {
		// Берём разрешение перед запуском. Остальные задачи ждут в этом цикле.
		if err := slots.acquire(ctx); err != nil {
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer slots.release()
			process(ctx, input)
		}()
	}

	wg.Wait()
	return ctx.Err()
}

func reuseSlotExample() {
	ctx := context.Background()
	slots := make(semaphore, 1)

	_ = slots.acquire(ctx) // У Background нет дедлайна и сигнала отмены.
	fmt.Println("First task acquired the only slot")
	slots.release()

	_ = slots.acquire(ctx)
	fmt.Println("Second task reused the slot")
	slots.release()
}

func batchExample() {
	var mu sync.Mutex
	total := 0
	numbers := []int{1, 2, 3, 4, 5}
	err := runLimited(context.Background(), numbers, 2, func(_ context.Context, number int) {
		squared := number * number
		mu.Lock()
		total += squared
		mu.Unlock()
	})
	fmt.Printf("Sum of squares: %d; error: %v\n", total, err)
}

func deadlineExample() {
	slots := make(semaphore, 1)
	_ = slots.acquire(context.Background())
	defer slots.release()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	// Место останется занятым до выхода из функции.
	fmt.Println("Waiting for the occupied slot:", slots.acquire(ctx))
}

func cancelBatchExample() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := 0
	err := runLimited(ctx, []int{1, 2, 3}, 1, func(_ context.Context, number int) {
		started++
		fmt.Println("Started task:", number)
		cancel()
	})
	// При возврате из runLimited все обработчики уже завершили работу.
	fmt.Printf("Started %d of 3; error: %v\n", started, err)
}

func invalidLimitExample() {
	err := runLimited(context.Background(), []int{1}, 0, func(context.Context, int) {
		fmt.Println("Unexpected task")
	})
	fmt.Println("Configuration:", err)
}

func main() {
	fmt.Println("=== 1. Reuse a slot ===")
	reuseSlotExample()
	fmt.Println("\n=== 2. Limit a batch ===")
	batchExample()
	fmt.Println("\n=== 3. Deadline waiting for a slot ===")
	deadlineExample()
	fmt.Println("\n=== 4. Cancel a batch ===")
	cancelBatchExample()
	fmt.Println("\n=== 5. Reject an invalid limit ===")
	invalidLimitExample()
}
