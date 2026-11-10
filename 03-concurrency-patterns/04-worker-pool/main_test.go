package main

import (
	"testing"
)

func TestWorkerPool(t *testing.T) {
	const workers = 3
	const tasks = 20

	jobs := make(chan Task, tasks)
	for i := 1; i <= tasks; i++ {
		jobs <- Task{ID: i}
	}
	close(jobs)

	results := make(chan Result)
	workerPool(jobs, results, workers)

	seen := make(map[int]int)
	for result := range results {
		seen[result.TaskID]++

		if result.WorkerID < 1 || result.WorkerID > workers {
			t.Fatalf("worker id %d out of range 1..%d", result.WorkerID, workers)
		}
	}

	// Каждая задача должна быть выполнена ровно один раз.
	if len(seen) != tasks {
		t.Fatalf("got %d unique tasks, want %d", len(seen), tasks)
	}

	for id, count := range seen {
		if count != 1 {
			t.Fatalf("task %d executed %d times, want 1", id, count)
		}
	}
}
