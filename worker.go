package work

import (
	"fmt"
	"math/rand"
	"reflect"

	"github.com/gomodule/redigo/redis"
)

const fetchKeysPerJobType = 6

var sleepBackoffsInMilliseconds = []int64{0, 10, 100, 1000, 5000}

type terminateOp func(conn redis.Conn)

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

func (w *worker) fetchJob() (*Job, error) {
	// resort queues
	// NOTE: could optimize this to only resort every second, or something.
	w.sampler.sample()
	numKeys := len(w.sampler.samples) * fetchKeysPerJobType
	scriptArgs := make([]interface{}, 0, numKeys+1)

	for _, s := range w.sampler.samples {
		scriptArgs = append(scriptArgs, s.redisJobs, s.redisJobsInProg, s.redisJobsPaused, s.redisJobsLock, s.redisJobsLockInfo, s.redisJobsMaxConcurrency) // KEYS[1-6 * N]
	}
	scriptArgs = append(scriptArgs, w.poolID) // ARGV[1]
	conn := w.pool.Get()
	defer conn.Close()

	values, err := redis.Values(w.redisFetchScript.Do(conn, scriptArgs...))
	if err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(values) != 3 {
		return nil, fmt.Errorf("need 3 elements back")
	}

	rawJSON, ok := values[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("response message not bytes")
	}

	dequeuedFrom, ok := values[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("response queue not bytes")
	}

	inProgQueue, ok := values[2].([]byte)
	if !ok {
		return nil, fmt.Errorf("response in prog not bytes")
	}

	job, err := newJob(rawJSON, dequeuedFrom, inProgQueue)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (w *worker) getAndDeleteUniqueJob(job *Job) *Job {
	var uniqueKey string
	var err error
	if job.UniqueKey != "" {
		uniqueKey = job.UniqueKey
	} else { // For jobs put in queue prior to this change. In the future this can be deleted as there will always be a UniqueKey
		uniqueKey, err = redisKeyUniqueJob(w.namespace, job.Name, job.Args)
		if err != nil {
			logError("worker.delete_unique_job.key", err)
			return nil
		}
	}

	conn := w.pool.Get()
	defer conn.Close()

	rawJSON, err := redis.Bytes(conn.Do("GET", uniqueKey))
	if err != nil {
		logError("worker.delete_unique_job.get", err)
		return nil
	}

	_, err = conn.Do("DEL", uniqueKey)
	if err != nil {
		logError("worker.delete_unique_job.del", err)
		return nil
	}

	// Previous versions did not support updated arguments and just set key to 1, so in these cases we should do nothing.
	// In the future this can be deleted, as we will always be getting arguments from here
	if string(rawJSON) == "1" {
		return nil
	}

	// The job pulled off the queue was just a placeholder with no args, so replace it
	jobWithArgs, err := newJob(rawJSON, job.dequeuedFrom, job.inProgQueue)
	if err != nil {
		logError("worker.delete_unique_job.updated_job", err)
		return nil
	}
	return jobWithArgs
}

// Default algorithm returns a fastly increasing unboundedly fashion backoff counter.
func defaultBackoffCalculator(job *Job) int64 {
	fails := job.Fails
	return (fails * fails * fails * fails) + 15 + (rand.Int63n(30) * (fails + 1))
}
