package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Main: started")

	go sayHello()

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Main: finished")
}

func sayHello() {
	fmt.Println("Goroutine: hello!")
}
