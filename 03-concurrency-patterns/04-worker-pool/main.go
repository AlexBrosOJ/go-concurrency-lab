package main

import (
	"fmt"
	"sync"
	"time"
)

// Task — задача, которую выполняют воркеры.
type Task struct {
	ID int
}

// Result — результат выполнения задачи.
type Result struct {
	TaskID   int
	WorkerID int
}

// workerPool запускает n воркеров, которые читают задачи из канала jobs
// и отправляют результаты в канал results. Когда канал jobs закрывается
// и все воркеры завершают работу, закрывается канал results.
//
// Важно: results закрывает одна горoutine после wg.Wait(), поэтому
// воркеры не могут отправить результат после закрытия канала.
func workerPool(jobs <-chan Task, results chan<- Result, n int) {
	var wg sync.WaitGroup

	for i := 1; i <= n; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for task := range jobs {
				// Имитируем полезную работу.
				time.Sleep(10 * time.Millisecond)

				results <- Result{
					TaskID:   task.ID,
					WorkerID: id,
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
}

func main() {
	const workers = 3
	const tasks = 10

	jobs := make(chan Task)
	results := make(chan Result, tasks)

	workerPool(jobs, results, workers)

	// Отправляем задачи в отдельной горутине, чтобы не блокироваться,
	// пока пул ещё не начал их читать.
	go func() {
		defer close(jobs)

		for i := 1; i <= tasks; i++ {
			jobs <- Task{ID: i}
		}
	}()

	// Собираем все результаты. Цикл завершится, когда закроется results.
	sum := 0
	for result := range results {
		fmt.Printf("задача %d выполнена воркером %d\n", result.TaskID, result.WorkerID)
		sum += result.TaskID
	}

	fmt.Printf("сумма ID всех задач: %d\n", sum)
}
