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
	//var mu sync.Mutex

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			//defer mu.Unlock()

			//mu.Lock()
			for j := 0; j < incrementsEach; j++ {
				counter++
			}
		}()
	}

	wg.Wait()

	fmt.Printf("Counter: %d\n", counter)
	fmt.Printf("Expected: %d\n", workers*incrementsEach)
}
