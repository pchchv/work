package work

import (
	"github.com/gomodule/redigo/redis"
)

type requeuer struct {
	namespace          string
	pool               *redis.Pool
	redisRequeueScript *redis.Script
	redisRequeueArgs   []interface{}
	stopChan           chan struct{}
	doneStoppingChan   chan struct{}
	drainChan          chan struct{}
	doneDrainingChan   chan struct{}
}
