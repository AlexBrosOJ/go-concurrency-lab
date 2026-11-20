package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// generate — источник. Слайс читается без копирования: не меняй его элементы
// до завершения производителя. Владелец out — только эта горутина.
func generate(ctx context.Context, wg *sync.WaitGroup, numbers []int) <-chan int {
	out := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for _, number := range numbers {
			if ctx.Err() != nil {
				return
			}
			select {
			case out <- number:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// square — преобразование: одному входному числу соответствует один квадрат.
// Отмена учитывается и при ожидании входа, и при ожидании потребителя выхода.
func square(ctx context.Context, wg *sync.WaitGroup, in <-chan int) <-chan int {
	out := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for {
			if ctx.Err() != nil {
				return
			}
			var number int
			select {
			case <-ctx.Done():
				return
			case value, ok := <-in:
				if !ok {
					return
				}
				number = value
			}
			result := number * number
			select {
			case out <- result:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// onlyEven — фильтр: нечётные значения не передаются дальше.
// Ни одна стадия не закрывает чужой вход.
func onlyEven(ctx context.Context, wg *sync.WaitGroup, in <-chan int) <-chan int {
	out := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for {
			if ctx.Err() != nil {
				return
			}
			var number int
			select {
			case <-ctx.Done():
				return
			case value, ok := <-in:
				if !ok {
					return
				}
				number = value
			}
			if number%2 != 0 {
				continue
			}
			select {
			case out <- number:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// pipeline строит все стадии до запуска Wait: регистрации задач уже завершены.
// finished закрывается после выхода всех трёх стадий, включая источник.
func pipeline(ctx context.Context, numbers []int) (values <-chan int, finished <-chan struct{}) {
	var wg sync.WaitGroup
	source := generate(ctx, &wg, numbers)
	squared := square(ctx, &wg, source)
	values = onlyEven(ctx, &wg, squared)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return values, done
}

func sourceExample() {
	var wg sync.WaitGroup
	for number := range generate(context.Background(), &wg, []int{1, 2, 3}) {
		fmt.Println("Source:", number)
	}
	wg.Wait()
}

func transformExample() {
	var wg sync.WaitGroup
	ctx := context.Background()
	for number := range square(ctx, &wg, generate(ctx, &wg, []int{-3, 0, 2})) {
		fmt.Println("Squared:", number)
	}
	wg.Wait()
}

func completePipelineExample() {
	values, finished := pipeline(context.Background(), []int{1, 2, 3, 4, 5, 6})
	for value := range values {
		fmt.Println("Even square:", value)
	}
	<-finished
	fmt.Println("All stages finished")
}

func reusableStageExample() {
	var wg sync.WaitGroup
	ctx := context.Background()
	source := generate(ctx, &wg, []int{1, 2, 3})
	squares := square(ctx, &wg, source)
	for value := range square(ctx, &wg, squares) {
		fmt.Println("Fourth power:", value)
	}
	wg.Wait()
}

func noMatchesExample() {
	values, finished := pipeline(context.Background(), []int{1, 3, 5})
	count := 0
	for range values {
		count++
	}
	<-finished
	fmt.Println("No matching results:", count)
}

func earlyExitExample() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	values, finished := pipeline(ctx, []int{2, 4, 6, 8, 10})
	if value, ok := <-values; ok {
		fmt.Println("Take one:", value)
	}
	cancel() // Дальше values не читаем: все отправки должны уметь отменяться.
	<-finished
	fmt.Println("Early exit: all stages stopped")
}

func unreadOutputExample() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, finished := pipeline(ctx, []int{2, 4, 6, 8})
	cancel()
	<-finished
	fmt.Println("Unread output: all stages stopped")
}

func deadlineExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	var wg sync.WaitGroup
	// У nil-входа никогда не будет данных. Спасти ожидание может ветка ctx.Done().
	out := square(ctx, &wg, nil)
	for range out {
		fmt.Println("Unexpected value")
	}
	wg.Wait()
	fmt.Println("Idle stage:", ctx.Err())
}

func main() {
	fmt.Println("=== 1. A source stage ===")
	sourceExample()
	fmt.Println("\n=== 2. A transformation stage ===")
	transformExample()
	fmt.Println("\n=== 3. Source -> square -> filter ===")
	completePipelineExample()
	fmt.Println("\n=== 4. Reuse a stage ===")
	reusableStageExample()
	fmt.Println("\n=== 5. No matching values ===")
	noMatchesExample()
	fmt.Println("\n=== 6. Consumer stops early ===")
	earlyExitExample()
	fmt.Println("\n=== 7. Consumer never reads ===")
	unreadOutputExample()
	fmt.Println("\n=== 8. Deadline while waiting for input ===")
	deadlineExample()
}
