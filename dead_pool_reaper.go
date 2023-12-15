package work

import "time"

const (
	deadTime          = 10 * time.Second // 2 x heartbeat
	reapPeriod        = 10 * time.Minute
	reapJitterSecs    = 30
	requeueKeysPerJob = 4
)
