package work

type observationKind int

type observation struct {
	kind observationKind
	// These fields always need to be set
	jobName string
	jobID   string
	// These need to be set when starting a job
	startedAt int64
	arguments map[string]interface{}
	// If we're done w/ the job, err will indicate the success/failure of it
	err error // nil: success. not nil: the error we got when running the job
	// If this is a checkin, set these.
	checkin   string
	checkinAt int64
}
