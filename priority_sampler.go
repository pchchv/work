package work

type sampleItem struct {
	priority uint
	// payload:
	redisJobs               string
	redisJobsInProg         string
	redisJobsPaused         string
	redisJobsLock           string
	redisJobsLockInfo       string
	redisJobsMaxConcurrency string
}

type prioritySampler struct {
	sum     uint
	samples []sampleItem
}

func (s *prioritySampler) add(
	priority uint,
	redisJobs,
	redisJobsInProg,
	redisJobsPaused,
	redisJobsLock,
	redisJobsLockInfo,
	redisJobsMaxConcurrency string) {
	sample := sampleItem{
		priority:                priority,
		redisJobs:               redisJobs,
		redisJobsInProg:         redisJobsInProg,
		redisJobsPaused:         redisJobsPaused,
		redisJobsLock:           redisJobsLock,
		redisJobsLockInfo:       redisJobsLockInfo,
		redisJobsMaxConcurrency: redisJobsMaxConcurrency,
	}
	s.samples = append(s.samples, sample)
	s.sum += priority
}
