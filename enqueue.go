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
