package work

// Q is a shortcut to easily specify arguments for jobs when enqueueing them.
// Example: e.Enqueue("send_email", work.Q{"addr": "test@example.com", "track": true})
type Q map[string]interface{}

// Job represents a job.
type Job struct {
	// Inputs when making a new job
	ID         string                 `json:"id"`
	Name       string                 `json:"name,omitempty"`
	Args       map[string]interface{} `json:"args"`
	Unique     bool                   `json:"unique,omitempty"`
	UniqueKey  string                 `json:"unique_key,omitempty"`
	EnqueuedAt int64                  `json:"t"`
	// Inputs when retrying
	Fails        int64  `json:"fails,omitempty"` // number of times this job has failed
	LastErr      string `json:"err,omitempty"`
	FailedAt     int64  `json:"failed_at,omitempty"`
	rawJSON      []byte
	argError     error
	observer     *observer
	inProgQueue  []byte
	dequeuedFrom []byte
}
