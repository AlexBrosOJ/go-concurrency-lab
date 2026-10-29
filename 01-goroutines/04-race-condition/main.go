package main

import (
	"fmt"
	"sync"
)

func main() {
	const (
		workers        = 5
		incrementsEach = 10000
	)

	var counter int
	var wg sync.WaitGroup

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < incrementsEach; j++ {
				counter++
			}
		}()
	}

	wg.Wait()

	fmt.Printf("Counter: %d\n", counter)
	fmt.Printf("Expected: %d\n", workers*incrementsEach)
}
