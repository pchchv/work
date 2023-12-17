package work

import (
	"sync"

	"github.com/gomodule/redigo/redis"
)

// Enqueuer can enqueue jobs.
type Enqueuer struct {
	Namespace             string // eg, "myapp-work"
	Pool                  *redis.Pool
	queuePrefix           string // eg, "myapp-work:jobs:"
	knownJobs             map[string]int64
	enqueueUniqueScript   *redis.Script
	enqueueUniqueInScript *redis.Script
	mtx                   sync.RWMutex
}

// NewEnqueuer creates a new enqueuer with
// the specified Redis namespace and Redis pool.
func NewEnqueuer(namespace string, pool *redis.Pool) *Enqueuer {
	if pool == nil {
		panic("NewEnqueuer needs a non-nil *redis.Pool")
	}
	return &Enqueuer{
		Namespace:             namespace,
		Pool:                  pool,
		queuePrefix:           redisKeyJobsPrefix(namespace),
		knownJobs:             make(map[string]int64),
		enqueueUniqueScript:   redis.NewScript(2, redisLuaEnqueueUnique),
		enqueueUniqueInScript: redis.NewScript(2, redisLuaEnqueueUniqueIn),
	}
}
