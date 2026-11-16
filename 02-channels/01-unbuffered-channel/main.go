package main

import "fmt"

// sendValue возвращается только после того, как получатель примет value.
// Канал должен быть небуферизированным; закрывает его вызывающий код при необходимости.
func sendValue(out chan<- int, value int) {
	out <- value
}

func singleValueExample() {
	values := make(chan int)
	finished := make(chan struct{})

	fmt.Println("Main: waiting for a value")
	go func() {
		sendValue(values, 42)
		// Получение значения ещё не означает завершения всей функции отправителя.
		close(finished)
	}()

	value := <-values
	fmt.Println("Main: received", value)
	<-finished
	fmt.Println("Main: sender finished")
}

// sendValues сохраняет порядок одного отправителя. Канал не закрывается:
// в этом примере получатель заранее знает количество значений.
func sendValues(out chan<- int, values []int) {
	for _, value := range values {
		out <- value
	}
}

// squareOnce обслуживает ровно один запрос. Канал ответа тоже небуферизированный.
func squareOnce(requests <-chan int, responses chan<- int) {
	value := <-requests
	responses <- value * value
}

func senderWaitsExample() {
	values := make(chan int)
	// Получателей ещё нет: выполнить отправку сейчас невозможно.
	select {
	case values <- 10:
		fmt.Println("Unexpected: send succeeded")
	default:
		fmt.Println("Send is not ready: no receiver")
	}
	finished := make(chan struct{})
	go func() {
		sendValue(values, 10)
		close(finished)
	}()
	fmt.Println("Received:", <-values)
	<-finished
}

func receiverWaitsExample() {
	values := make(chan int)
	// Отправителей ещё нет: канал не возвращает ноль автоматически.
	select {
	case value := <-values:
		fmt.Println("Unexpected value:", value)
	default:
		fmt.Println("Receive is not ready: no sender")
	}
	finished := make(chan struct{})
	go func() {
		sendValue(values, 0)
		close(finished)
	}()
	fmt.Println("Received a real zero:", <-values)
	<-finished
}

func sequenceExample() {
	values := make(chan int)
	numbers := []int{10, 20, 30}
	finished := make(chan struct{})
	go func() {
		sendValues(values, numbers)
		close(finished)
	}()
	// Количество получений совпадает с количеством отправок.
	for i := 0; i < len(numbers); i++ {
		fmt.Println("Next:", <-values)
	}
	<-finished
}

func requestReplyExample() {
	requests := make(chan int)
	responses := make(chan int)
	finished := make(chan struct{})
	go func() {
		squareOnce(requests, responses)
		close(finished)
	}()
	requests <- 7
	// Отправка запроса завершена, но ответ нужно получить отдельно.
	fmt.Println("Square:", <-responses)
	<-finished
}

func acknowledgementExample() {
	jobs := make(chan string)
	processed := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		job := <-jobs
		fmt.Println("Processed:", job)
		// Это отправка подтверждения после обработки, а не close.
		processed <- struct{}{}
		close(finished)
	}()
	jobs <- "learn channels"
	// Не печатаем из main сразу после отправки: обработка ещё могла не закончиться.
	<-processed
	fmt.Println("Main: processing confirmed")
	<-finished
}

func main() {
	fmt.Println("=== 1. One value and sender completion ===")
	singleValueExample()
	fmt.Println("\n=== 2. Send needs a receiver ===")
	senderWaitsExample()
	fmt.Println("\n=== 3. Receive needs a sender ===")
	receiverWaitsExample()
	fmt.Println("\n=== 4. Several values in order ===")
	sequenceExample()
	fmt.Println("\n=== 5. Request and reply ===")
	requestReplyExample()
	fmt.Println("\n=== 6. Processing acknowledgement ===")
	acknowledgementExample()
}
