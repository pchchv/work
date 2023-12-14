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
