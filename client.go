package work

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

// ScheduledJob represents a job in the scheduled queue.
type ScheduledJob struct {
	RunAt int64 `json:"run_at"`
	*Job
}

// WorkerPoolHeartbeat represents the heartbeat from a worker pool.
// WorkerPool's write a heartbeat every 5 seconds so we know they're alive and includes config information.
type WorkerPoolHeartbeat struct {
	WorkerPoolID string   `json:"worker_pool_id"`
	StartedAt    int64    `json:"started_at"`
	HeartbeatAt  int64    `json:"heartbeat_at"`
	JobNames     []string `json:"job_names"`
	Concurrency  uint     `json:"concurrency"`
	Host         string   `json:"host"`
	Pid          int      `json:"pid"`
	WorkerIDs    []string `json:"worker_ids"`
}

// WorkerObservation represents the latest observation taken from a worker.
// The observation indicates whether the worker is busy processing a job,
// and if so, information about that job.
type WorkerObservation struct {
	WorkerID string `json:"worker_id"`
	IsBusy   bool   `json:"is_busy"`
	// If IsBusy:
	JobName   string `json:"job_name"`
	JobID     string `json:"job_id"`
	StartedAt int64  `json:"started_at"`
	ArgsJSON  string `json:"args_json"`
	Checkin   string `json:"checkin"`
	CheckinAt int64  `json:"checkin_at"`
}

// Queue represents a queue that holds jobs with the same name.
// It indicates their name, count, and latency (in seconds).
// Latency is a measurement of how long ago the next job to be processed was enqueued.
type Queue struct {
	JobName string `json:"job_name"`
	Count   int64  `json:"count"`
	Latency int64  `json:"latency"`
}

// RetryJob represents a job in the retry queue.
type RetryJob struct {
	RetryAt int64 `json:"retry_at"`
	*Job
}

// DeadJob represents a job in the dead queue.
type DeadJob struct {
	DiedAt int64 `json:"died_at"`
	*Job
}

type jobScore struct {
	JobBytes []byte
	Score    int64
	job      *Job
}

// Client implements all of the functionality of the web UI.
// It can be used to inspect the status of a running cluster and retry dead jobs.
type Client struct {
	namespace string
	pool      *redis.Pool
}

// NewClient creates a new Client with the specified redis namespace and connection pool.
func NewClient(namespace string, pool *redis.Pool) *Client {
	return &Client{
		namespace: namespace,
		pool:      pool,
	}
}

// WorkerPoolHeartbeats queries Redis and returns all WorkerPoolHeartbeat's it finds
// (even for those worker pools which don't have a current heartbeat).
func (c *Client) WorkerPoolHeartbeats() ([]*WorkerPoolHeartbeat, error) {
	conn := c.pool.Get()
	defer conn.Close()

	workerPoolsKey := redisKeyWorkerPools(c.namespace)
	workerPoolIDs, err := redis.Strings(conn.Do("SMEMBERS", workerPoolsKey))
	if err != nil {
		return nil, err
	}
	sort.Strings(workerPoolIDs)

	for _, wpid := range workerPoolIDs {
		key := redisKeyHeartbeat(c.namespace, wpid)
		conn.Send("HGETALL", key)
	}

	if err := conn.Flush(); err != nil {
		logError("worker_pool_statuses.flush", err)
		return nil, err
	}

	heartbeats := make([]*WorkerPoolHeartbeat, 0, len(workerPoolIDs))
	for _, wpid := range workerPoolIDs {
		vals, err := redis.Strings(conn.Receive())
		if err != nil {
			logError("worker_pool_statuses.receive", err)
			return nil, err
		}

		heartbeat := &WorkerPoolHeartbeat{
			WorkerPoolID: wpid,
		}

		for i := 0; i < len(vals)-1; i += 2 {
			var err error
			key := vals[i]
			value := vals[i+1]

			switch key {
			case "heartbeat_at":
				heartbeat.HeartbeatAt, err = strconv.ParseInt(value, 10, 64)
			case "started_at":
				heartbeat.StartedAt, err = strconv.ParseInt(value, 10, 64)
			case "job_names":
				heartbeat.JobNames = strings.Split(value, ",")
				sort.Strings(heartbeat.JobNames)
			case "concurrency":
				var vv uint64
				vv, err = strconv.ParseUint(value, 10, 0)
				heartbeat.Concurrency = uint(vv)
			case "host":
				heartbeat.Host = value
			case "pid":
				var vv int64
				vv, err = strconv.ParseInt(value, 10, 0)
				heartbeat.Pid = int(vv)
			case "worker_ids":
				heartbeat.WorkerIDs = strings.Split(value, ",")
				sort.Strings(heartbeat.WorkerIDs)
			}
			if err != nil {
				logError("worker_pool_statuses.parse", err)
				return nil, err
			}
		}
		heartbeats = append(heartbeats, heartbeat)
	}
	return heartbeats, nil
}

// WorkerObservations returns all of the WorkerObservation's it finds for all worker pools' workers.
func (c *Client) WorkerObservations() ([]*WorkerObservation, error) {
	conn := c.pool.Get()
	defer conn.Close()

	hbs, err := c.WorkerPoolHeartbeats()
	if err != nil {
		logError("worker_observations.worker_pool_heartbeats", err)
		return nil, err
	}

	var workerIDs []string
	for _, hb := range hbs {
		workerIDs = append(workerIDs, hb.WorkerIDs...)
	}

	for _, wid := range workerIDs {
		key := redisKeyWorkerObservation(c.namespace, wid)
		conn.Send("HGETALL", key)
	}

	if err := conn.Flush(); err != nil {
		logError("worker_observations.flush", err)
		return nil, err
	}

	observations := make([]*WorkerObservation, 0, len(workerIDs))

	for _, wid := range workerIDs {
		vals, err := redis.Strings(conn.Receive())
		if err != nil {
			logError("worker_observations.receive", err)
			return nil, err
		}

		ob := &WorkerObservation{
			WorkerID: wid,
		}

		for i := 0; i < len(vals)-1; i += 2 {
			var err error
			key := vals[i]
			ob.IsBusy = true
			value := vals[i+1]

			switch key {
			case "job_name":
				ob.JobName = value
			case "job_id":
				ob.JobID = value
			case "started_at":
				ob.StartedAt, err = strconv.ParseInt(value, 10, 64)
			case "args":
				ob.ArgsJSON = value
			case "checkin":
				ob.Checkin = value
			case "checkin_at":
				ob.CheckinAt, err = strconv.ParseInt(value, 10, 64)
			}
			if err != nil {
				logError("worker_observations.parse", err)
				return nil, err
			}
		}
		observations = append(observations, ob)
	}
	return observations, nil
}
