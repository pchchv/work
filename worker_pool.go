package work

import (
	"reflect"

	"github.com/gomodule/redigo/redis"
)

// You may provide your own backoff function for retrying failed jobs or use the builtin one.
// Returns the number of seconds to wait until the next attempt.
//
// The builtin backoff calculator provides an exponentially increasing wait function.
type BackoffCalculator func(job *Job) int64

// JobOptions can be passed to JobWithOptions.
type JobOptions struct {
	Priority       uint              // Priority from 1 to 10000
	MaxFails       uint              // 1: send straight to dead (unless SkipDead)
	SkipDead       bool              // If true, don't send failed jobs to the dead queue when retries are exhausted.
	MaxConcurrency uint              // Max number of jobs to keep in flight (default is 0, meaning no max)
	Backoff        BackoffCalculator // If not set, uses the default backoff algorithm
}

// GenericHandler is a job handler without any custom context.
type GenericHandler func(*Job) error

// GenericMiddlewareHandler is a middleware without any custom context.
type GenericMiddlewareHandler func(*Job, NextMiddlewareFunc) error

// NextMiddlewareFunc is a function type
// (whose instances are named 'next')
// that you call to advance to the next middleware.
type NextMiddlewareFunc func() error

type middlewareHandler struct {
	IsGeneric                bool
	DynamicMiddleware        reflect.Value
	GenericMiddlewareHandler GenericMiddlewareHandler
}

// WorkerPoolOptions can be passed to NewWorkerPoolWithOptions.
type WorkerPoolOptions struct {
	SleepBackoffs []int64 // Sleep backoffs in milliseconds
}

type jobType struct {
	Name string
	JobOptions
	IsGeneric      bool
	GenericHandler GenericHandler
	DynamicHandler reflect.Value
}

func (jt *jobType) calcBackoff(j *Job) int64 {
	if jt.Backoff == nil {
		return defaultBackoffCalculator(j)
	}
	return jt.Backoff(j)
}

// WorkerPool represents a pool of workers.
// It forms the primary API of gocraft/work.
// WorkerPools provide the public API of gocraft/work.
// You can attach jobs and middlware to them.
// You can start and stop them.
// Based on their concurrency setting,
// they'll spin up N worker goroutines.
type WorkerPool struct {
	workerPoolID     string
	concurrency      uint
	namespace        string // eg, "myapp-work"
	pool             *redis.Pool
	sleepBackoffs    []int64
	contextType      reflect.Type
	jobTypes         map[string]*jobType
	middleware       []*middlewareHandler
	started          bool
	periodicJobs     []*periodicJob
	workers          []*worker
	heartbeater      *workerPoolHeartbeater
	retrier          *requeuer
	scheduler        *requeuer
	deadPoolReaper   *deadPoolReaper
	periodicEnqueuer *periodicEnqueuer
}

// NewWorkerPoolWithOptions creates a new worker pool as per the NewWorkerPool function, but permits you to specify
// additional options such as sleep backoffs.
func NewWorkerPoolWithOptions(ctx interface{}, concurrency uint, namespace string, pool *redis.Pool, workerPoolOpts WorkerPoolOptions) *WorkerPool {
	if pool == nil {
		panic("NewWorkerPool needs a non-nil *redis.Pool")
	}

	ctxType := reflect.TypeOf(ctx)
	validateContextType(ctxType)
	wp := &WorkerPool{
		workerPoolID:  makeIdentifier(),
		concurrency:   concurrency,
		namespace:     namespace,
		pool:          pool,
		sleepBackoffs: workerPoolOpts.SleepBackoffs,
		contextType:   ctxType,
		jobTypes:      make(map[string]*jobType),
	}

	for i := uint(0); i < wp.concurrency; i++ {
		w := newWorker(wp.namespace, wp.workerPoolID, wp.pool, wp.contextType, nil, wp.jobTypes, wp.sleepBackoffs)
		wp.workers = append(wp.workers, w)
	}
	return wp
}

// NewWorkerPool creates a new worker pool.
// ctx should be a struct literal whose type will be used for middleware and handlers.
// concurrency specifies how many workers to spin up - each worker can process jobs concurrently.
func NewWorkerPool(ctx interface{}, concurrency uint, namespace string, pool *redis.Pool) *WorkerPool {
	return NewWorkerPoolWithOptions(ctx, concurrency, namespace, pool, WorkerPoolOptions{})
}

// validateContextType will panic if context is invalid.
func validateContextType(ctxType reflect.Type) {
	if ctxType.Kind() != reflect.Struct {
		panic("work: Context needs to be a struct type")
	}
}
