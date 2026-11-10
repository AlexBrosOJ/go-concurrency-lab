package main

import (
	"sync/atomic"
	"testing"
)

func TestRateLimiterRespectsLimit(t *testing.T) {
	const limit = 4

	var inFlight atomic.Int64
	var maxInFlight atomic.Int64

	jobs := make([]int, 50)
	for i := range jobs {
		jobs[i] = i
	}

	rateLimiter(jobs, limit, func(int) {
		current := inFlight.Add(1)

		for {
			previous := maxInFlight.Load()
			if current <= previous || maxInFlight.CompareAndSwap(previous, current) {
				break
			}
		}

		inFlight.Add(-1)
	})

	if got := maxInFlight.Load(); got > limit {
		t.Fatalf("max concurrent goroutines %d exceeds limit %d", got, limit)
	}
}

func TestRateLimiterProcessesAllJobs(t *testing.T) {
	const jobs = 100

	var processed atomic.Int64

	rateLimiter(makeRange(jobs), 5, func(int) {
		processed.Add(1)
	})

	if got := processed.Load(); got != jobs {
		t.Fatalf("processed %d jobs, want %d", got, jobs)
	}
}

func makeRange(n int) []int {
	values := make([]int, n)
	for i := range values {
		values[i] = i
	}
	return values
}
