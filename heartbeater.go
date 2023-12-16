package work

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

type workerPoolHeartbeater struct {
	workerPoolID     string
	namespace        string // eg, "myapp-work"
	pool             *redis.Pool
	beatPeriod       time.Duration
	concurrency      uint
	jobNames         string
	startedAt        int64
	pid              int
	hostname         string
	workerIDs        string
	stopChan         chan struct{}
	doneStoppingChan chan struct{}
}
