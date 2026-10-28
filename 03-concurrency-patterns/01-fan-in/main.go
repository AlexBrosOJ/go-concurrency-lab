package main

import (
	"fmt"
	"sync"
)

func fanIn(channels ...<-chan int) <-chan int {
	out := make(chan int)

	var wg sync.WaitGroup
	wg.Add(len(channels))

	for _, ch := range channels {
		go func(ch <-chan int) {
			defer wg.Done()

			for value := range ch {
				out <- value
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	first := make(chan int)
	second := make(chan int)

	go func() {
		defer close(first)

		for _, value := range []int{1, 3, 5} {
			first <- value
		}
	}()

	go func() {
		defer close(second)

		for _, value := range []int{2, 4, 6} {
			second <- value
		}
	}()

	for value := range fanIn(first, second) {
		fmt.Println(value)
	}
}
