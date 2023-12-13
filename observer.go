package work

import "github.com/gomodule/redigo/redis"

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

// An observer observes a single worker. Each worker has its own observer.
type observer struct {
	namespace string
	workerID  string
	pool      *redis.Pool
	// nil: worker isn't doing anything that we know of
	// not nil: the last started observation that we received on the channel.
	// if we get an checkin, we'll just update the existing observation
	currentStartedObservation *observation
	// version of the data that we wrote to redis.
	// each observation we get, we'll update version. When we flush it to redis, we'll update lastWrittenVersion.
	// This will keep us from writing to redis unless necessary
	version            int64
	lastWrittenVersion int64
	observationsChan   chan *observation
	stopChan           chan struct{}
	doneStoppingChan   chan struct{}
	drainChan          chan struct{}
	doneDrainingChan   chan struct{}
}
