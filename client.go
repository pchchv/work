package work

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
