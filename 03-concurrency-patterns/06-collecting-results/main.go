package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Result struct {
	Index int
	Value int
	Err   error
}

type Entry struct {
	Value    int
	Err      error
	Received bool
}

// runTasks запускает по горутине на вход. Подходит для небольших партий.
func runTasks(ctx context.Context, inputs []int, work func(context.Context, int) (int, error)) <-chan Result {
	out := make(chan Result)
	var wg sync.WaitGroup
	wg.Add(len(inputs))
	for index, input := range inputs {
		go func() {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			value, err := work(ctx, input)
			select {
			case out <- Result{Index: index, Value: value, Err: err}:
			case <-ctx.Done():
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// inInputOrder принимает результаты из runTasks и раскладывает их по индексам.
// Received отличает отсутствующий результат от доставленного нуля или ошибки.
func inInputOrder(size int, results <-chan Result) []Entry {
	entries := make([]Entry, size)
	for result := range results {
		entries[result.Index] = Entry{Value: result.Value, Err: result.Err, Received: true}
	}
	return entries
}

func doublePositive(_ context.Context, input int) (int, error) {
	if input < 0 {
		return 0, errors.New("negative input")
	}
	return input * 2, nil
}

func completionOrderExample() {
	ctx := context.Background()
	firstMayFinish := make(chan struct{})
	results := runTasks(ctx, []int{10, 20}, func(_ context.Context, input int) (int, error) {
		if input == 10 {
			<-firstMayFinish
		}
		return input, nil
	})
	first := <-results
	fmt.Printf("First arrival: index=%d value=%d\n", first.Index, first.Value)
	// Разрешаем первой входной задаче закончить только после получения второй.
	close(firstMayFinish)
	for result := range results {
		fmt.Printf("Next arrival: index=%d value=%d\n", result.Index, result.Value)
	}
}

func inputOrderExample() {
	inputs := []int{3, 0, -4, 3}
	entries := inInputOrder(len(inputs), runTasks(context.Background(), inputs, doublePositive))
	for index, entry := range entries {
		fmt.Printf("Index %d: value=%d error=%v received=%t\n", index, entry.Value, entry.Err, entry.Received)
	}
}

func emptyExample() {
	results := runTasks(context.Background(), nil, doublePositive)
	fmt.Println("Entries:", len(inInputOrder(0, results)))
}

func partialResultsExample() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results := runTasks(ctx, []int{0, 2, 3}, func(ctx context.Context, input int) (int, error) {
		if input != 0 {
			<-ctx.Done()
		}
		return input, ctx.Err()
	})
	// Первый результат уже доставлен. Прекращаем читать и отменяем остаток.
	first := <-results
	cancel()
	// В этой демонстрации сохраняем снимок на момент решения об отмене.
	// Дочитываем канал только для ожидания выхода горутин.
	for range results {
	}
	fmt.Printf("Kept index %d: value=%d; other entries not retained\n", first.Index, first.Value)
}

func main() {
	fmt.Println("=== 1. Arrival order ===")
	completionOrderExample()
	fmt.Println("\n=== 2. Input order with errors ===")
	inputOrderExample()
	fmt.Println("\n=== 3. Empty batch ===")
	emptyExample()
	fmt.Println("\n=== 4. Keep a partial result ===")
	partialResultsExample()
}
