package main

import (
	"fmt"
	"sync"
)

// fanOut распределяет значения из одного входного канала
// между n горутин-обработчиков. Каждая горoutine применяет функцию process
// к полученным значениям и отправляет результат в общий выходной канал.
//
// Когда входной канал закрывается, все обработчики завершают работу,
// после чего закрывается выходной канал.
func fanOut[T, R any](in <-chan T, n int, process func(T) R) <-chan R {
	out := make(chan R)

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			for value := range in {
				out <- process(value)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	in := make(chan int)

	go func() {
		defer close(in)

		for _, value := range []int{1, 2, 3, 4, 5, 6} {
			in <- value
		}
	}()

	// Четыре обработчика возводят каждое число в квадрат.
	// Порядок результатов не гарантирован — это особенность паттерна.
	for result := range fanOut(in, 4, func(value int) int {
		return value * value
	}) {
		fmt.Println(result)
	}
}
