package work

import "math/rand"

// Default algorithm returns a fastly increasing unboundedly fashion backoff counter.
func defaultBackoffCalculator(job *Job) int64 {
	fails := job.Fails
	return (fails * fails * fails * fails) + 15 + (rand.Int63n(30) * (fails + 1))
}
