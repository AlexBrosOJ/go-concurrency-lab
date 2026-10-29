package main

import (
	"fmt"
	"time"
)

func printNumbers(workerID int) {
	for i := 1; i <= 5; i++ {
		fmt.Printf("Worker %d: %d\n", workerID, i)
		time.Sleep(100 * time.Microsecond)
	}
}

func main() {
	fmt.Println("Main: statred")

	go printNumbers(1)
	go printNumbers(2)
	go printNumbers(3)

	time.Sleep(600 * time.Millisecond)

	fmt.Println("Main: finished")
}
