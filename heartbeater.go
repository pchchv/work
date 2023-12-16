package work

import (
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

const beatPeriod = 5 * time.Second

type workerPoolHeartbeater struct {
	workerPoolID     string
	namespace        string // eg, "myapp-work"
	pool             *redis.Pool
	beatPeriod       time.Duration
	concurrency      uint
	jobNames         string
	startedAt        int64
	pid              int
	hostname         string
	workerIDs        string
	stopChan         chan struct{}
	doneStoppingChan chan struct{}
}

func newWorkerPoolHeartbeater(
	namespace string,
	pool *redis.Pool,
	workerPoolID string,
	jobTypes map[string]*jobType,
	concurrency uint,
	workerIDs []string) *workerPoolHeartbeater {
	h := &workerPoolHeartbeater{
		workerPoolID:     workerPoolID,
		namespace:        namespace,
		pool:             pool,
		beatPeriod:       beatPeriod,
		concurrency:      concurrency,
		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),
	}

	jobNames := make([]string, 0, len(jobTypes))
	for k := range jobTypes {
		jobNames = append(jobNames, k)
	}
	sort.Strings(jobNames)
	h.jobNames = strings.Join(jobNames, ",")

	sort.Strings(workerIDs)
	h.workerIDs = strings.Join(workerIDs, ",")

	h.pid = os.Getpid()
	host, err := os.Hostname()
	if err != nil {
		logError("heartbeat.hostname", err)
		host = "hostname_errored"
	}

	h.hostname = host
	return h
}
