package work

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestDeadPoolReaper(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	conn := pool.Get()
	defer conn.Close()

	workerPoolsKey := redisKeyWorkerPools(ns)

	// Create redis data
	var err error
	err = conn.Send("SADD", workerPoolsKey, "1")
	assert.NoError(t, err)
	err = conn.Send("SADD", workerPoolsKey, "2")
	assert.NoError(t, err)
	err = conn.Send("SADD", workerPoolsKey, "3")
	assert.NoError(t, err)

	err = conn.Send("HMSET", redisKeyHeartbeat(ns, "1"),
		"heartbeat_at", time.Now().Unix(),
		"job_names", "type1,type2",
	)
	assert.NoError(t, err)

	err = conn.Send("HMSET", redisKeyHeartbeat(ns, "2"),
		"heartbeat_at", time.Now().Add(-1*time.Hour).Unix(),
		"job_names", "type1,type2",
	)
	assert.NoError(t, err)

	err = conn.Send("HMSET", redisKeyHeartbeat(ns, "3"),
		"heartbeat_at", time.Now().Add(-1*time.Hour).Unix(),
		"job_names", "type1,type2",
	)
	assert.NoError(t, err)
	err = conn.Flush()
	assert.NoError(t, err)

	// Test getting dead pool
	reaper := newDeadPoolReaper(ns, pool, []string{})
	deadPools, err := reaper.findDeadPools()
	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"2": {"type1", "type2"}, "3": {"type1", "type2"}}, deadPools)

	// Test requeueing jobs
	_, err = conn.Do("lpush", redisKeyJobsInProgress(ns, "2", "type1"), "foo")
	assert.NoError(t, err)
	_, err = conn.Do("incr", redisKeyJobsLock(ns, "type1"))
	assert.NoError(t, err)
	_, err = conn.Do("hincrby", redisKeyJobsLockInfo(ns, "type1"), "2", 1) // worker pool 2 has lock
	assert.NoError(t, err)

	// Ensure 0 jobs in jobs queue
	jobsCount, err := redis.Int(conn.Do("llen", redisKeyJobs(ns, "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 0, jobsCount)

	// Ensure 1 job in inprogress queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "2", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)

	// Reap
	err = reaper.reap()
	assert.NoError(t, err)

	// Ensure 1 jobs in jobs queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobs(ns, "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)

	// Ensure 0 job in inprogress queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "2", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 0, jobsCount)

	// Locks should get cleaned up
	assert.EqualValues(t, 0, getInt64(pool, redisKeyJobsLock(ns, "type1")))
	v, _ := conn.Do("HGET", redisKeyJobsLockInfo(ns, "type1"), "2")
	assert.Nil(t, v)
}
