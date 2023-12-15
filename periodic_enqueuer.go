package work

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron/v3"
)

const periodicEnqueuerSleep = 2 * time.Minute

type periodicJob struct {
	spec     string
	jobName  string
	schedule cron.Schedule
}

type scheduledPeriodicJob struct {
	scheduledAt      time.Time
	scheduledAtEpoch int64
	*periodicJob
}

type periodicEnqueuer struct {
	namespace             string
	pool                  *redis.Pool
	periodicJobs          []*periodicJob
	scheduledPeriodicJobs []*scheduledPeriodicJob
	stopChan              chan struct{}
	doneStoppingChan      chan struct{}
}

func newPeriodicEnqueuer(namespace string, pool *redis.Pool, periodicJobs []*periodicJob) *periodicEnqueuer {
	return &periodicEnqueuer{
		namespace:        namespace,
		pool:             pool,
		periodicJobs:     periodicJobs,
		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),
	}
}

func (pe *periodicEnqueuer) shouldEnqueue() bool {
	conn := pe.pool.Get()
	defer conn.Close()

	lastEnqueue, err := redis.Int64(conn.Do("GET", redisKeyLastPeriodicEnqueue(pe.namespace)))
	if err == redis.ErrNil {
		return true
	} else if err != nil {
		logError("periodic_enqueuer.should_enqueue", err)
		return true
	}

	return lastEnqueue < (nowEpochSeconds() - int64(periodicEnqueuerSleep/time.Minute))
}
