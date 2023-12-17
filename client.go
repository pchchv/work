package work

// ScheduledJob represents a job in the scheduled queue.
type ScheduledJob struct {
	RunAt int64 `json:"run_at"`
	*Job
}
