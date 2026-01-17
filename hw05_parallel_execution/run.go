package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 {
		return nil
	}

	if (n <= 0) || (m <= 0) {
		return ErrErrorsLimitExceeded
	}

	// set then error count exceeds or equal m
	stop := make(chan struct{})
	defer close(stop)

	active := make(chan struct{}, n)
	defer close(active)

	// current error count
	errorCount := 0

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

taskLoop:
	for i := range len(tasks) {
		select {
		case <-stop:
			break taskLoop

		case active <- struct{}{}:
			wg.Add(1)

			go func(task Task) {
				defer wg.Done()
				defer func() { <-active }()

				err := task()
				if err != nil {
					mu.Lock()
					defer func() { mu.Unlock() }()

					errorCount++

					if errorCount == m {
						stop <- struct{}{}
					}
				}
			}(tasks[i])
		}
	}
	wg.Wait()

	if errorCount >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
