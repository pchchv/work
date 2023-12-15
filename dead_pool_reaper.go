package work

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	deadTime          = 10 * time.Second // 2 x heartbeat
	reapPeriod        = 10 * time.Minute
	reapJitterSecs    = 30
	requeueKeysPerJob = 4
)

type deadPoolReaper struct {
	namespace        string
	pool             *redis.Pool
	deadTime         time.Duration
	reapPeriod       time.Duration
	curJobTypes      []string
	stopChan         chan struct{}
	doneStoppingChan chan struct{}
}
