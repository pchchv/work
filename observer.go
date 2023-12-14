package work

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
)

const (
	observerBufferSize                     = 1024
	observationKindStarted observationKind = iota
	observationKindDone
	observationKindCheckin
)

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

func newObserver(namespace string, pool *redis.Pool, workerID string) *observer {
	return &observer{
		namespace:        namespace,
		workerID:         workerID,
		pool:             pool,
		observationsChan: make(chan *observation, observerBufferSize),
		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),
		drainChan:        make(chan struct{}),
		doneDrainingChan: make(chan struct{}),
	}
}

func (o *observer) observeCheckin(jobName, jobID, checkin string) {
	o.observationsChan <- &observation{
		kind:      observationKindCheckin,
		jobName:   jobName,
		jobID:     jobID,
		checkin:   checkin,
		checkinAt: nowEpochSeconds(),
	}
}

func (o *observer) writeStatus(obv *observation) error {
	conn := o.pool.Get()
	defer conn.Close()
	key := redisKeyWorkerObservation(o.namespace, o.workerID)

	if obv == nil {
		if _, err := conn.Do("DEL", key); err != nil {
			return err
		}
	} else {
		// hash:
		// job_name -> obv.Name
		// job_id -> obv.jobID
		// started_at -> obv.startedAt
		// args -> json.Encode(obv.arguments)
		// checkin -> obv.checkin
		// checkin_at -> obv.checkinAt
		var argsJSON []byte
		if len(obv.arguments) == 0 {
			argsJSON = []byte("")
		} else {
			var err error
			argsJSON, err = json.Marshal(obv.arguments)
			if err != nil {
				return err
			}
		}

		args := make([]interface{}, 0, 13)
		args = append(args,
			key,
			"job_name", obv.jobName,
			"job_id", obv.jobID,
			"started_at", obv.startedAt,
			"args", argsJSON,
		)

		if (obv.checkin != "") && (obv.checkinAt > 0) {
			args = append(args,
				"checkin", obv.checkin,
				"checkin_at", obv.checkinAt,
			)
		}

		conn.Send("HMSET", args...)
		conn.Send("EXPIRE", key, 60*60*24)
		if err := conn.Flush(); err != nil {
			return err
		}
	}
	return nil
}
