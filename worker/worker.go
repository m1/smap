package worker

import (
	"sync"
)

type worker struct {
	id   int
	wg   *sync.WaitGroup
	jobs chan Job
}

func (w *worker) run() {
	defer w.wg.Done()
	for job := range w.jobs {
		job.Run()
	}
}
