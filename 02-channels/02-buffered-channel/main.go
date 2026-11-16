package main

import "fmt"

// sendValues отправляет значения в исходном порядке и не закрывает канал.
// Если буфер заполнен, отправитель ждёт получателя.
func sendValues(out chan<- int, values []int) {
	for _, value := range values {
		out <- value
	}
}

func capacityExample() {
	values := make(chan int, 2)
	fmt.Printf("Created: len=%d cap=%d\n", len(values), cap(values))
	// Две отправки помещаются в буфер даже без другой горутины.
	values <- 10
	values <- 20
	fmt.Printf("Filled: len=%d cap=%d\n", len(values), cap(values))
	fmt.Println("Received:", <-values)
	fmt.Printf("After receive: len=%d cap=%d\n", len(values), cap(values))
	fmt.Println("Received:", <-values)
}

func fifoExample() {
	values := make(chan string, 3)
	values <- "first"
	values <- "second"
	values <- "third"
	for i := 0; i < 3; i++ {
		fmt.Println("Next:", <-values)
	}
}

func fullBufferExample() {
	values := make(chan int, 2)
	values <- 10
	values <- 20
	// Нет получателя, а свободных мест уже нет: отправка не готова.
	select {
	case values <- 30:
		fmt.Println("Unexpected: send succeeded")
	default:
		fmt.Println("Buffer full: send is not ready")
	}
	fmt.Println("Freed one slot:", <-values)
	values <- 30 // Освободившееся место позволяет отправить без получателя.
	fmt.Println("Remaining:", <-values, <-values)
}

func emptyBufferExample() {
	values := make(chan int, 1)
	select {
	case value := <-values:
		fmt.Println("Unexpected value:", value)
	default:
		fmt.Println("Buffer empty: receive is not ready")
	}
	values <- 0
	fmt.Println("Received a real zero:", <-values)
}

func producerConsumerExample() {
	values := make(chan int, 2)
	numbers := []int{1, 2, 3, 4, 5}
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		sendValues(values, numbers)
	}()
	// Значений больше, чем мест в буфере. Получатель освобождает места,
	// поэтому отправитель может постепенно передать всю последовательность.
	for i := 0; i < len(numbers); i++ {
		fmt.Println("Consumed:", <-values)
	}
	<-finished
	fmt.Println("Producer finished")
}

func closeAndDrainExample() {
	values := make(chan int, 2)
	values <- 0
	values <- 42
	close(values)
	fmt.Println("Closed with buffered values:", len(values))
	// Закрытие не удаляет уже переданные значения.
	for value := range values {
		fmt.Println("Drained:", value)
	}
	value, ok := <-values
	fmt.Printf("After drain: value=%d ok=%t\n", value, ok)
}

func queuedIsNotProcessedExample() {
	jobs := make(chan string, 1)
	allowProcessing := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		// Явный барьер вместо задержки: обработка ещё не разрешена.
		<-allowProcessing
		job := <-jobs
		fmt.Println("Processed:", job)
	}()
	jobs <- "learn buffered channels"
	fmt.Println("Job queued; processing is still gated")
	select {
	case <-finished:
		fmt.Println("Unexpected: worker finished")
	default:
		fmt.Println("Worker has not finished")
	}
	close(allowProcessing)
	<-finished
	fmt.Println("Main: processing completed")
}

func main() {
	fmt.Println("=== 1. Capacity and length ===")
	capacityExample()
	fmt.Println("\n=== 2. FIFO order ===")
	fifoExample()
	fmt.Println("\n=== 3. Full buffer and a free slot ===")
	fullBufferExample()
	fmt.Println("\n=== 4. Empty buffer ===")
	emptyBufferExample()
	fmt.Println("\n=== 5. Producer and consumer ===")
	producerConsumerExample()
	fmt.Println("\n=== 6. Close and drain ===")
	closeAndDrainExample()
	fmt.Println("\n=== 7. Queued is not processed ===")
	queuedIsNotProcessedExample()
}
