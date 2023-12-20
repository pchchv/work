package work

import (
	"testing"

	"github.com/robfig/cron/v3"
)

func TestPeriodicEnqueuerSpawn(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	pe := newPeriodicEnqueuer(ns, pool, nil)
	pe.start()
	pe.stop()
}

func appendPeriodicJob(pjs []*periodicJob, spec, jobName string) []*periodicJob {
	p := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	sched, err := p.Parse(spec)
	if err != nil {
		panic(err)
	}

	pj := &periodicJob{jobName: jobName, spec: spec, schedule: sched}
	return append(pjs, pj)
}
