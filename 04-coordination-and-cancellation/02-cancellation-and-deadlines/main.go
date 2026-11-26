package main

import (
	"context"
	"fmt"
	"time"
)

// collectWithin создаёт один бюджет на весь поток и возвращает уже прочитанные данные.
func collectWithin(parent context.Context, timeout time.Duration, input <-chan int) ([]int, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	var values []int
	for {
		if err := ctx.Err(); err != nil {
			return values, err
		}
		select {
		case value, ok := <-input:
			if !ok {
				return values, ctx.Err()
			}
			values = append(values, value)
		case <-ctx.Done():
			return values, ctx.Err()
		}
	}
}

func closedInput(values ...int) <-chan int {
	input := make(chan int, len(values))
	for _, value := range values {
		input <- value
	}
	close(input)
	return input
}

func fullStreamExample() {
	values, err := collectWithin(context.Background(), time.Second, closedInput(4, 0, 4))
	fmt.Printf("Collected: %v; error: %v\n", values, err)
}

func exhaustedBudgetExample() {
	values, err := collectWithin(context.Background(), 0, nil)
	fmt.Printf("Collected: %v; error: %v\n", values, err)
}

func parentCancellationExample() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	values, err := collectWithin(ctx, time.Hour, closedInput(7))
	fmt.Printf("Collected: %v; error: %v\n", values, err)
}

func inheritedDeadlineExample() {
	parent, cancelParent := context.WithTimeout(context.Background(), time.Minute)
	defer cancelParent()
	parentDeadline, _ := parent.Deadline()
	child, cancelChild := context.WithDeadline(parent, parentDeadline.Add(time.Minute))
	defer cancelChild()
	childDeadline, _ := child.Deadline()
	fmt.Println("Child keeps the earlier parent deadline:", childDeadline.Equal(parentDeadline))
}

func liveTimeoutExample() {
	// Открытый канал без отправителя позволяет наблюдать реальное ожидание таймера.
	values, err := collectWithin(context.Background(), 10*time.Millisecond, make(chan int))
	fmt.Printf("Collected: %v; error: %v\n", values, err)
}

func main() {
	fmt.Println("=== 1. Read a finite stream ===")
	fullStreamExample()
	fmt.Println("\n=== 2. Budget already exhausted ===")
	exhaustedBudgetExample()
	fmt.Println("\n=== 3. Parent cancelled before reading ===")
	parentCancellationExample()
	fmt.Println("\n=== 4. Earlier parent deadline wins ===")
	inheritedDeadlineExample()
	fmt.Println("\n=== 5. Timeout waiting for data ===")
	liveTimeoutExample()
}
