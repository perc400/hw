package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func doTasks(waitGroup *sync.WaitGroup, once *sync.Once, errorsCount *int32, maxErrorsCount int, tasksChannel chan Task, doneChannel chan struct{}) {
	defer waitGroup.Done()

	for task := range tasksChannel {
		if task() != nil {
			atomic.AddInt32(errorsCount, 1)
			if *errorsCount >= int32(maxErrorsCount) {
				once.Do(func() { close(doneChannel) })
				return
			}
		}
	}
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}

	var wg sync.WaitGroup
	var once sync.Once
	var errorsCount int32

	tasksCh := make(chan Task)
	doneCh := make(chan struct{})

	for i := 0; i < n; i++ {
		wg.Add(1)
		go doTasks(&wg, &once, &errorsCount, m, tasksCh, doneCh)
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
