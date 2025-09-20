package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}

	var wg sync.WaitGroup
	var once sync.Once
	var errorsCount int64

	tasksCh := make(chan Task)
	doneCh := make(chan struct{})

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, once *sync.Once) {
			defer wg.Done()
			for task := range tasksCh {
				if task() != nil {
					atomic.AddInt64(&errorsCount, 1)
					if atomic.LoadInt64(&errorsCount) >= int64(m) {
						once.Do(func() { close(doneCh) })
						return
					}
				}
			}
		}(&wg, &once)
	}

	breakSending := false
	for _, task := range tasks {
		select {
		case <-doneCh:
			breakSending = true
		case tasksCh <- task:
		}
		if breakSending {
			break
		}
	}
	close(tasksCh)

	wg.Wait()

	select {
	case <-doneCh:
		return ErrErrorsLimitExceeded
	default:
		return nil
	}
}
