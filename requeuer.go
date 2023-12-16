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

func newRequeuer(namespace string, pool *redis.Pool, requeueKey string, jobNames []string) *requeuer {
	args := make([]interface{}, 0, len(jobNames)+2+2)
	args = append(args, requeueKey)              // KEY[1]
	args = append(args, redisKeyDead(namespace)) // KEY[2]
	for _, jobName := range jobNames {
		args = append(args, redisKeyJobs(namespace, jobName)) // KEY[3, 4, ...]
	}

	args = append(args, redisKeyJobsPrefix(namespace)) // ARGV[1]
	args = append(args, 0)                             // ARGV[2] -- NOTE: We're going to change this one on every call
	return &requeuer{
		namespace:          namespace,
		pool:               pool,
		redisRequeueScript: redis.NewScript(len(jobNames)+2, redisLuaZremLpushCmd),
		redisRequeueArgs:   args,
		stopChan:           make(chan struct{}),
		doneStoppingChan:   make(chan struct{}),
		drainChan:          make(chan struct{}),
		doneDrainingChan:   make(chan struct{}),
	}
}
