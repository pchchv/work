package work

import "github.com/robfig/cron/v3"

type periodicJob struct {
	spec     string
	jobName  string
	schedule cron.Schedule
}
