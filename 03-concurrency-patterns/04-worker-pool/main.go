package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
)

type Job struct {
	ID   int
	Text string
}

type Result struct {
	ID    int
	Words int
	Err   error
}

type processor func(context.Context, Job) (int, error)

// startPool обрабатывает партию задач. Слайс batch можно менять после закрытия out.
// При раннем выходе потребитель отменяет контекст и дочитывает out до закрытия.
func startPool(ctx context.Context, batch []Job, workers, queueSize int, process processor) (<-chan Result, error) {
	if workers <= 0 || queueSize < 0 {
		return nil, errors.New("workers must be positive and queue size nonnegative")
	}

	jobs := make(chan Job, queueSize)
	out := make(chan Result)
	var wg sync.WaitGroup
	wg.Add(workers + 1)

	// Канал jobs закрывает только производитель.
	go func() {
		defer wg.Done()
		defer close(jobs)
		for _, job := range batch {
			select {
			case jobs <- job:
			case <-ctx.Done():
				return
			}
		}
	}()

	for range workers {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok || ctx.Err() != nil {
						return
					}
					words, err := process(ctx, job)
					select {
					case out <- Result{ID: job.ID, Words: words, Err: err}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out, nil
}

func countWords(_ context.Context, job Job) (int, error) {
	words := strings.Fields(job.Text)
	if len(words) == 0 {
		return 0, errors.New("empty text")
	}
	return len(words), nil
}

func printBatch(workers, queueSize int, batch []Job) {
	out, err := startPool(context.Background(), batch, workers, queueSize, countWords)
	if err != nil {
		fmt.Println(err)
		return
	}
	var results []Result
	for result := range out {
		results = append(results, result)
	}
	// В примере ID уникальны. Сортируем результаты для воспроизводимого вывода.
	slices.SortFunc(results, func(a, b Result) int {
		return cmp.Compare(a.ID, b.ID)
	})
	for _, result := range results {
		if result.Err != nil {
			fmt.Printf("Job %d: %v\n", result.ID, result.Err)
			continue
		}
		fmt.Printf("Job %d: %d words\n", result.ID, result.Words)
	}
	fmt.Printf("Pool finished: %d results\n", len(results))
}

func earlyExitExample() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// При одном обработчике первый результат соответствует первой задаче.
	out, err := startPool(ctx, []Job{{1, "keep this"}, {2, "discard the rest"}}, 1, 0, countWords)
	if err != nil {
		fmt.Println(err)
		return
	}
	if result, ok := <-out; ok {
		fmt.Printf("Kept job %d: %d words\n", result.ID, result.Words)
	}
	cancel()
	for range out {
		// Готовый результат может прийти одновременно с отменой.
	}
	fmt.Println("Pool stopped after cancellation")
}

func main() {
	batch := []Job{{3, "three short words"}, {1, "one"}, {2, "two words"}, {4, "  "}}

	fmt.Println("=== 1. One worker, no queue buffer ===")
	printBatch(1, 0, batch[:3])

	fmt.Println("\n=== 2. Two workers, buffered queue, one job error ===")
	printBatch(2, 2, batch)

	fmt.Println("\n=== 3. Empty batch ===")
	printBatch(3, 1, nil)

	fmt.Println("\n=== 4. Consumer stops early ===")
	earlyExitExample()

	fmt.Println("\n=== 5. Invalid configuration ===")
	printBatch(0, 1, batch)
}
