package work

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
)

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

// ArgString returns j.Args[key] typed to a string.
// If the key is missing or of the wrong type, it sets an argument error
// on the job. This function is meant to be used in
// the body of a job handling function while extracting arguments,
// followed by a single call to j.ArgError().
func (j *Job) ArgString(key string) string {
	v, ok := j.Args[key]
	if ok {
		typedV, ok := v.(string)
		if ok {
			return typedV
		}
		j.argError = typecastError("string", key, v)
	} else {
		j.argError = missingKeyError("string", key)
	}
	return ""
}

func typecastError(jsonType, key string, v interface{}) error {
	actualType := reflect.TypeOf(v)
	return fmt.Errorf("looking for a %s in job.Arg[%s] but value wasn't right type: %v(%v)", jsonType, key, actualType, v)
}

func missingKeyError(jsonType, key string) error {
	return fmt.Errorf("looking for a %s in job.Arg[%s] but key wasn't found", jsonType, key)
}

func isIntKind(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64
}

func isUintKind(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64
}

func isFloatKind(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Float32 || k == reflect.Float64
}
