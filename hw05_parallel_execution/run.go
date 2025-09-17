package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func sendTasks(tasks []Task, tasksChannel chan Task, doneChannel chan struct{}) {
	for _, task := range tasks {
		select {
		case <-doneChannel:
			close(tasksChannel)
			return
		case tasksChannel <- task:
		}
	}
	close(tasksChannel)
}

func countErrors(errorsChannel chan error, maxErrorsCount int, doneChannel chan struct{}) {
	errorsCount := 0
	for err := range errorsChannel {
		if err != nil {
			errorsCount++
			if errorsCount >= maxErrorsCount {
				close(doneChannel)
				return
			}
		}
	}
}

func doTasks(waitGroup *sync.WaitGroup, tasksChannel chan Task, errorsChannel chan error, doneChannel chan struct{}) {
	defer waitGroup.Done()
	for {
		select {
		case <-doneChannel:
			return
		case task, ok := <-tasksChannel:
			if !ok {
				return
			}
			if task != nil {
				select {
				case <-doneChannel:
					return
				case errorsChannel <- task():
				}
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

	tasksCh := make(chan Task)
	errorsCh := make(chan error)
	doneCh := make(chan struct{})

	for i := 0; i < n; i++ {
		wg.Add(1)
		go doTasks(&wg, tasksCh, errorsCh, doneCh)
	}

	go countErrors(errorsCh, m, doneCh)

	go sendTasks(tasks, tasksCh, doneCh)

	wg.Wait()
	close(errorsCh)

	select {
	case <-doneCh:
		return ErrErrorsLimitExceeded
	default:
		return nil
	}
}
