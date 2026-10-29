package main

import (
	"fmt"
	"sync"
	"time"
)

func printNumbers(workerID int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 1; i <= 5; i++ {
		fmt.Printf("Worker %d: %d\n", workerID, i)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("Worker %d: finished\n", workerID)
}

func main() {
	fmt.Println("Main: started")

	var wg sync.WaitGroup

	for workerID := 1; workerID <= 3; workerID++ {
		wg.Add(1)
		go printNumbers(workerID, &wg)
	}

	wg.Wait()

	fmt.Println("Main: finished")
}
