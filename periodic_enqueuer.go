package work

import (
	"time"

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
