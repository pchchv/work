package work

import "time"

var nowMock int64

func nowEpochSeconds() int64 {
	if nowMock != 0 {
		return nowMock
	}
	return time.Now().Unix()
}
