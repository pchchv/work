package work

import "testing"

func TestPeriodicEnqueuerSpawn(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	pe := newPeriodicEnqueuer(ns, pool, nil)
	pe.start()
	pe.stop()
}
