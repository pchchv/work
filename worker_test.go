package work

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestWorkerBasics(t *testing.T) {
	var arg1 float64
	var arg2 float64
	var arg3 float64
	pool := newTestPool(":6379")
	ns := "work"
	job1 := "job1"
	job2 := "job2"
	job3 := "job3"

	cleanKeyspace(ns, pool)

	jobTypes := make(map[string]*jobType)
	jobTypes[job1] = &jobType{
		Name:       job1,
		JobOptions: JobOptions{Priority: 1},
		IsGeneric:  true,
		GenericHandler: func(job *Job) error {
			arg1 = job.Args["a"].(float64)
			return nil
		},
	}
	jobTypes[job2] = &jobType{
		Name:       job2,
		JobOptions: JobOptions{Priority: 1},
		IsGeneric:  true,
		GenericHandler: func(job *Job) error {
			arg2 = job.Args["a"].(float64)
			return nil
		},
	}
	jobTypes[job3] = &jobType{
		Name:       job3,
		JobOptions: JobOptions{Priority: 1},
		IsGeneric:  true,
		GenericHandler: func(job *Job) error {
			arg3 = job.Args["a"].(float64)
			return nil
		},
	}

	enqueuer := NewEnqueuer(ns, pool)
	_, err := enqueuer.Enqueue(job1, Q{"a": 1})
	assert.Nil(t, err)
	_, err = enqueuer.Enqueue(job2, Q{"a": 2})
	assert.Nil(t, err)
	_, err = enqueuer.Enqueue(job3, Q{"a": 3})
	assert.Nil(t, err)

	w := newWorker(ns, "1", pool, tstCtxType, nil, jobTypes, nil)
	w.start()
	w.drain()
	w.stop()

	// make sure the jobs ran (side effect of setting these variables to the job arguments)
	assert.EqualValues(t, 1.0, arg1)
	assert.EqualValues(t, 2.0, arg2)
	assert.EqualValues(t, 3.0, arg3)

	// nothing in retries or dead
	assert.EqualValues(t, 0, zsetSize(pool, redisKeyRetry(ns)))
	assert.EqualValues(t, 0, zsetSize(pool, redisKeyDead(ns)))

	// nothing in the queues or in-progress queues
	assert.EqualValues(t, 0, listSize(pool, redisKeyJobs(ns, job1)))
	assert.EqualValues(t, 0, listSize(pool, redisKeyJobs(ns, job2)))
	assert.EqualValues(t, 0, listSize(pool, redisKeyJobs(ns, job3)))
	assert.EqualValues(t, 0, listSize(pool, redisKeyJobsInProgress(ns, "1", job1)))
	assert.EqualValues(t, 0, listSize(pool, redisKeyJobsInProgress(ns, "1", job2)))
	assert.EqualValues(t, 0, listSize(pool, redisKeyJobsInProgress(ns, "1", job3)))

	// nothing in the worker status
	h := readHash(pool, redisKeyWorkerObservation(ns, w.workerID))
	assert.EqualValues(t, 0, len(h))
}

func newTestPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxActive:   10,
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		Wait: true,
	}
}

func cleanKeyspace(namespace string, pool *redis.Pool) {
	conn := pool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", namespace+"*"))
	if err != nil {
		panic("could not get keys: " + err.Error())
	}

	for _, k := range keys {
		if _, err := conn.Do("DEL", k); err != nil {
			panic("could not del: " + err.Error())
		}
	}
}

func zsetSize(pool *redis.Pool, key string) int64 {
	conn := pool.Get()
	defer conn.Close()

	v, err := redis.Int64(conn.Do("ZCARD", key))
	if err != nil {
		panic("could not get ZSET size: " + err.Error())
	}
	return v
}

func listSize(pool *redis.Pool, key string) int64 {
	conn := pool.Get()
	defer conn.Close()

	v, err := redis.Int64(conn.Do("LLEN", key))
	if err != nil {
		panic("could not get list length: " + err.Error())
	}
	return v
}

func deleteQueue(pool *redis.Pool, namespace, jobName string) {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", redisKeyJobs(namespace, jobName), redisKeyJobsInProgress(namespace, "1", jobName))
	if err != nil {
		panic("could not delete queue: " + err.Error())
	}
}

func deleteRetryAndDead(pool *redis.Pool, namespace string) {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", redisKeyRetry(namespace), redisKeyDead(namespace))
	if err != nil {
		panic("could not delete retry/dead queue: " + err.Error())
	}
}
