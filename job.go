package work

import "encoding/json"

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

func newJob(rawJSON, dequeuedFrom, inProgQueue []byte) (*Job, error) {
	var job Job
	if err := json.Unmarshal(rawJSON, &job); err != nil {
		return nil, err
	}

	job.rawJSON = rawJSON
	job.dequeuedFrom = dequeuedFrom
	job.inProgQueue = inProgQueue
	return &job, nil
}

func (j *Job) serialize() ([]byte, error) {
	return json.Marshal(j)
}

// setArg sets a single named argument on the job.
func (j *Job) setArg(key string, val interface{}) {
	if j.Args == nil {
		j.Args = make(map[string]interface{})
	}
	j.Args[key] = val
}

func (j *Job) failed(err error) {
	j.Fails++
	j.LastErr = err.Error()
	j.FailedAt = nowEpochSeconds()
}

// Checkin will update the status of the executing job to the specified messages.
// This message is visible within the web UI.
// This is useful for indicating some sort of progress on very long running jobs.
// For instance, on a job that has to process a million records over the course of an hour,
// the job could call Checkin with the current job number every 10k jobs.
func (j *Job) Checkin(msg string) {
	if j.observer != nil {
		j.observer.observeCheckin(j.Name, j.ID, msg)
	}
}
