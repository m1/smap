package worker

import (
	"sync"
)

// Pool defines a pool of workers, with a max size
// and stores a jobs queue
type Pool struct {
	size int
	jobs chan Job

	workersWg *sync.WaitGroup
	jobsWg    *sync.WaitGroup
}

// NewPool passes back a new pool of workers
func NewPool(size int) *Pool {
	pool := &Pool{
		size:      size,
		jobs:      make(chan Job),
		workersWg: &sync.WaitGroup{},
		jobsWg:    &sync.WaitGroup{},
	}

	return pool
}

// Start creates the workers and running their jobs
func (p *Pool) Start() {
	for i := 0; i < p.size; i++ {
		p.workersWg.Add(1)
		w := worker{
			id:   i,
			wg:   p.workersWg,
			jobs: p.jobs,
		}

		go w.run()
	}

	go func() {
		p.workersWg.Wait()
	}()
}

// AddJob adds a job to the queue for the worker pool
// to consume
func (p *Pool) AddJob(job Job) {
	p.jobsWg.Add(1)
	go func() {
		p.jobs <- job
		p.jobsWg.Done()
	}()
}

// Close indicates that the jobs have finished and tries
// to close the jobs queue and waits for the workers to finish
func (p *Pool) Close() {
	p.jobsWg.Wait()
	close(p.jobs)
}
