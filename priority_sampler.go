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
