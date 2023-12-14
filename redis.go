package work

import "fmt"

func redisNamespacePrefix(namespace string) string {
	l := len(namespace)
	if (l > 0) && (namespace[l-1] != ':') {
		namespace = namespace + ":"
	}
	return namespace
}

func redisKeyWorkerObservation(namespace, workerID string) string {
	return redisNamespacePrefix(namespace) + "worker:" + workerID
}

// returns "<namespace>:jobs:"
// so that we can just append the job name and be good to go
func redisKeyJobsPrefix(namespace string) string {
	return redisNamespacePrefix(namespace) + "jobs:"
}

func redisKeyJobs(namespace, jobName string) string {
	return redisKeyJobsPrefix(namespace) + jobName
}

func redisKeyJobsInProgress(namespace, poolID, jobName string) string {
	return fmt.Sprintf("%s:%s:inprogress", redisKeyJobs(namespace, jobName), poolID)
}

func redisKeyJobsLock(namespace, jobName string) string {
	return redisKeyJobs(namespace, jobName) + ":lock"
}

func redisKeyJobsPaused(namespace, jobName string) string {
	return redisKeyJobs(namespace, jobName) + ":paused"
}

func redisKeyJobsLockInfo(namespace, jobName string) string {
	return redisKeyJobs(namespace, jobName) + ":lock_info"
}
