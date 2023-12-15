package work

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron/v3"
)

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
