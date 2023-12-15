package work

import (
	"math/rand"
	"reflect"

	"github.com/gomodule/redigo/redis"
)

const fetchKeysPerJobType = 6

var sleepBackoffsInMilliseconds = []int64{0, 10, 100, 1000, 5000}

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

func newWorker(namespace string, poolID string, pool *redis.Pool, contextType reflect.Type, middleware []*middlewareHandler, jobTypes map[string]*jobType, sleepBackoffs []int64) *worker {
	workerID := makeIdentifier()
	ob := newObserver(namespace, pool, workerID)

	if len(sleepBackoffs) == 0 {
		sleepBackoffs = sleepBackoffsInMilliseconds
	}

	w := &worker{
		workerID:      workerID,
		poolID:        poolID,
		namespace:     namespace,
		pool:          pool,
		contextType:   contextType,
		sleepBackoffs: sleepBackoffs,

		observer: ob,

		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),

		drainChan:        make(chan struct{}),
		doneDrainingChan: make(chan struct{}),
	}

	w.updateMiddlewareAndJobTypes(middleware, jobTypes)

	return w
}

// note: can't be called while the thing is started.
func (w *worker) updateMiddlewareAndJobTypes(middleware []*middlewareHandler, jobTypes map[string]*jobType) {
	w.middleware = middleware
	sampler := prioritySampler{}
	for _, jt := range jobTypes {
		sampler.add(jt.Priority,
			redisKeyJobs(w.namespace, jt.Name),
			redisKeyJobsInProgress(w.namespace, w.poolID, jt.Name),
			redisKeyJobsPaused(w.namespace, jt.Name),
			redisKeyJobsLock(w.namespace, jt.Name),
			redisKeyJobsLockInfo(w.namespace, jt.Name),
			redisKeyJobsConcurrency(w.namespace, jt.Name))
	}
	w.sampler = sampler
	w.jobTypes = jobTypes
	w.redisFetchScript = redis.NewScript(len(jobTypes)*fetchKeysPerJobType, redisLuaFetchJob)
}

func (w *worker) start() {
	go w.loop()
	go w.observer.start()
}

func (w *worker) stop() {
	w.stopChan <- struct{}{}
	<-w.doneStoppingChan
	w.observer.drain()
	w.observer.stop()
}

func (w *worker) drain() {
	w.drainChan <- struct{}{}
	<-w.doneDrainingChan
	w.observer.drain()
}

// Default algorithm returns a fastly increasing unboundedly fashion backoff counter.
func defaultBackoffCalculator(job *Job) int64 {
	fails := job.Fails
	return (fails * fails * fails * fails) + 15 + (rand.Int63n(30) * (fails + 1))
}
