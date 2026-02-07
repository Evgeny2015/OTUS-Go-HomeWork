package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

type RunData struct {
	errorCount    int
	maxErrorCount int
	wg            sync.WaitGroup
	mu            sync.RWMutex
	done          chan struct{}
	task          chan Task
}

// IncError increment error count and check if we should stop.
func (r *RunData) IncError() bool {
	r.mu.Lock()
	r.errorCount++
	shouldStop := r.errorCount >= r.maxErrorCount
	r.mu.Unlock()

	// Signal to stop if we've reached the error limit
	if shouldStop {
		select {
		case r.done <- struct{}{}:
		default:
		}
	}

	return shouldStop
}

// ShouldStop check error count.
func (r *RunData) ShouldStop() bool {
	r.mu.RLock()
	shouldStop := r.errorCount >= r.maxErrorCount
	r.mu.RUnlock()

	return shouldStop
}

func CreateRunData(maxErrorCount int) *RunData {
	return &RunData{
		maxErrorCount: maxErrorCount,
		wg:            sync.WaitGroup{},
		mu:            sync.RWMutex{},
		done:          make(chan struct{}),
		task:          make(chan Task),
	}
}

func (r *RunData) Close() {
	close(r.done)
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 {
		return nil
	}

	if (n <= 0) || (m <= 0) {
		return ErrErrorsLimitExceeded
	}

	runData := CreateRunData(m)
	defer runData.Close()

	// start n worker goroutines
	for range n {
		runData.wg.Add(1)
		go worker(runData)
	}

	// start task distributor
	runData.wg.Add(1)
	go distributor(runData, tasks)

	runData.wg.Wait()

	if runData.errorCount >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}

// worker processes tasks from the task channel.
func worker(runData *RunData) {
	defer runData.wg.Done()

	for taskInstance := range runData.task {
		// Check if we should stop processing
		select {
		case <-runData.done:
			return
		default:
		}

		// Execute the task
		err := taskInstance()
		if err != nil {
			// Increment error count and check if we should stop
			if runData.IncError() {
				break
			}
		}
	}
}

// distributor sends tasks to workers until the error limit is reached or all tasks are sent.
func distributor(runData *RunData, tasks []Task) {
	defer runData.wg.Done()
	defer close(runData.task)

	for i := range tasks {
		// Check if we should stop
		select {
		case <-runData.done:
			return
		default:
		}

		if runData.ShouldStop() {
			return
		}

		// Send task to worker
		select {
		case runData.task <- tasks[i]:
		case <-runData.done:
			return
		}
	}
}
