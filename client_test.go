package work

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestContext struct{}

func TestClientWorkerPoolHeartbeats(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	wp := NewWorkerPool(TestContext{}, 10, ns, pool)
	wp.Job("wat", func(job *Job) error { return nil })
	wp.Job("bob", func(job *Job) error { return nil })
	wp.Start()

	wp2 := NewWorkerPool(TestContext{}, 11, ns, pool)
	wp2.Job("foo", func(job *Job) error { return nil })
	wp2.Job("bar", func(job *Job) error { return nil })
	wp2.Start()

	time.Sleep(20 * time.Millisecond)

	client := NewClient(ns, pool)

	hbs, err := client.WorkerPoolHeartbeats()
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(hbs))
	if len(hbs) == 2 {
		var hbwp, hbwp2 *WorkerPoolHeartbeat

		if wp.workerPoolID == hbs[0].WorkerPoolID {
			hbwp = hbs[0]
			hbwp2 = hbs[1]
		} else {
			hbwp = hbs[1]
			hbwp2 = hbs[0]
		}

		assert.Equal(t, wp.workerPoolID, hbwp.WorkerPoolID)
		assert.EqualValues(t, uint(10), hbwp.Concurrency)
		assert.Equal(t, []string{"bob", "wat"}, hbwp.JobNames)
		assert.Equal(t, wp.workerIDs(), hbwp.WorkerIDs)

		assert.Equal(t, wp2.workerPoolID, hbwp2.WorkerPoolID)
		assert.EqualValues(t, uint(11), hbwp2.Concurrency)
		assert.Equal(t, []string{"bar", "foo"}, hbwp2.JobNames)
		assert.Equal(t, wp2.workerIDs(), hbwp2.WorkerIDs)
	}

	wp.Stop()
	wp2.Stop()

	hbs, err = client.WorkerPoolHeartbeats()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(hbs))
}

func TestClientWorkerObservations(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	enqueuer := NewEnqueuer(ns, pool)
	_, err := enqueuer.Enqueue("wat", Q{"a": 1, "b": 2})
	assert.Nil(t, err)
	_, err = enqueuer.Enqueue("foo", Q{"a": 3, "b": 4})
	assert.Nil(t, err)

	wp := NewWorkerPool(TestContext{}, 10, ns, pool)
	wp.Job("wat", func(job *Job) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	wp.Job("foo", func(job *Job) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	wp.Start()

	time.Sleep(10 * time.Millisecond)

	client := NewClient(ns, pool)
	observations, err := client.WorkerObservations()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(observations))

	watCount := 0
	fooCount := 0
	for _, ob := range observations {
		if ob.JobName == "foo" {
			fooCount++
			assert.True(t, ob.IsBusy)
			assert.Equal(t, `{"a":3,"b":4}`, ob.ArgsJSON)
			assert.True(t, (nowEpochSeconds()-ob.StartedAt) <= 3)
			assert.True(t, ob.JobID != "")
		} else if ob.JobName == "wat" {
			watCount++
			assert.True(t, ob.IsBusy)
			assert.Equal(t, `{"a":1,"b":2}`, ob.ArgsJSON)
			assert.True(t, (nowEpochSeconds()-ob.StartedAt) <= 3)
			assert.True(t, ob.JobID != "")
		} else {
			assert.False(t, ob.IsBusy)
		}
		assert.True(t, ob.WorkerID != "")
	}
	assert.Equal(t, 1, watCount)
	assert.Equal(t, 1, fooCount)

	wp.Stop()

	observations, err = client.WorkerObservations()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(observations))
}
