package main

import (
	"fmt"
	"sync"
)

// fanOut распределяет числа между workers обработчиками и возвращает их квадраты.
// workers должен быть положительным. Вызывающий закрывает in и дочитывает результат.
func fanOut(in <-chan int, workers int) <-chan int {
	if workers <= 0 {
		panic("fanOut: workers must be positive")
	}
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for value := range in {
				out <- value * value
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
	jobs := make(chan int)
	go func() {
		defer close(jobs)
		for _, value := range []int{1, 2, 3, 4, 5, 6} {
			jobs <- value
		}
	}()
	for result := range fanOut(jobs, 3) {
		fmt.Println(result)
	}
}
