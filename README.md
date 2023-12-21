# work [![Go Reference](https://pkg.go.dev/badge/github.com/pchchv/work.svg)](https://pkg.go.dev/github.com/pchchv/work)

**work** provides the ability to enqueue and process background jobs in Go.  
Jobs are durable and supported by Redis.

* Fast and efficient.
* Reliable — does not lose jobs even if the process crashes.
* Middleware on jobs — good for metrics, logging, etc.
* If a job fails, it will be retried a certain number of times.
* Scheduling jobs for the future.
* Enqueue unique jobs so that only one job with a given name/arguments is in the queue at once.
* Web interface to manage failed jobs and monitor system performance.
* Periodic queuing of jobs on a cron-like schedule.
* Pause/unpause jobs and control concurrency within and across processes.
