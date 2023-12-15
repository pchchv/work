package work

import (
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	deadTime          = 10 * time.Second // 2 x heartbeat
	reapPeriod        = 10 * time.Minute
	reapJitterSecs    = 30
	requeueKeysPerJob = 4
)

type deadPoolReaper struct {
	namespace        string
	pool             *redis.Pool
	deadTime         time.Duration
	reapPeriod       time.Duration
	curJobTypes      []string
	stopChan         chan struct{}
	doneStoppingChan chan struct{}
}

func newDeadPoolReaper(namespace string, pool *redis.Pool, curJobTypes []string) *deadPoolReaper {
	return &deadPoolReaper{
		namespace:        namespace,
		pool:             pool,
		deadTime:         deadTime,
		reapPeriod:       reapPeriod,
		curJobTypes:      curJobTypes,
		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),
	}
}

func (r *deadPoolReaper) findDeadPools() (map[string][]string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	workerPoolsKey := redisKeyWorkerPools(r.namespace)

	workerPoolIDs, err := redis.Strings(conn.Do("SMEMBERS", workerPoolsKey))
	if err != nil {
		return nil, err
	}

	deadPools := map[string][]string{}
	for _, workerPoolID := range workerPoolIDs {
		heartbeatKey := redisKeyHeartbeat(r.namespace, workerPoolID)
		heartbeatAt, err := redis.Int64(conn.Do("HGET", heartbeatKey, "heartbeat_at"))
		if err == redis.ErrNil {
			// heartbeat expired, save dead pool and use cur set of jobs from reaper
			deadPools[workerPoolID] = []string{}
			continue
		}
		if err != nil {
			return nil, err
		}

		// Check that last heartbeat was long enough ago to consider the pool dead
		if time.Unix(heartbeatAt, 0).Add(r.deadTime).After(time.Now()) {
			continue
		}

		jobTypesList, err := redis.String(conn.Do("HGET", heartbeatKey, "job_names"))
		if err == redis.ErrNil {
			continue
		}
		if err != nil {
			return nil, err
		}

		deadPools[workerPoolID] = strings.Split(jobTypesList, ",")
	}

	return deadPools, nil
}
