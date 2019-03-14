package worker

// Job is the interface for the jobs that the workers
// can execute. Once a worker selects a job, it runs
// the function `Run`
type Job interface {
	Run()
}
