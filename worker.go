package work

import (
	"math/rand"
	"reflect"

	"github.com/gomodule/redigo/redis"
)

type worker struct {
	workerID         string
	poolID           string
	namespace        string
	pool             *redis.Pool
	jobTypes         map[string]*jobType
	sleepBackoffs    []int64
	middleware       []*middlewareHandler
	contextType      reflect.Type
	redisFetchScript *redis.Script
	sampler          prioritySampler
	*observer
	stopChan         chan struct{}
	doneStoppingChan chan struct{}
	drainChan        chan struct{}
	doneDrainingChan chan struct{}
}

// Default algorithm returns a fastly increasing unboundedly fashion backoff counter.
func defaultBackoffCalculator(job *Job) int64 {
	fails := job.Fails
	return (fails * fails * fails * fails) + 15 + (rand.Int63n(30) * (fails + 1))
}
