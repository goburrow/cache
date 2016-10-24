package cache

import "sync/atomic"

// Stats is statistics about performance of a cache.
type Stats struct {
	HitCount         uint64
	MissCount        uint64
	LoadSuccessCount uint64
	LoadErrorCount   uint64
	EvictionCount    uint64
}

// AddHitCount increases HitCount atomically.
func (s *Stats) AddHitCount(delta uint64) {
	atomic.AddUint64(&s.HitCount, delta)
}

// AddMissCount increases MissCount atomically.
func (s *Stats) AddMissCount(delta uint64) {
	atomic.AddUint64(&s.MissCount, delta)
}

// AddLoadSuccessCount increases LoadSuccessCount atomically.
func (s *Stats) AddLoadSuccessCount(delta uint64) {
	atomic.AddUint64(&s.LoadSuccessCount, delta)
}

// AddLoadErrorCount increases LoadErrorCount atomically.
func (s *Stats) AddLoadErrorCount(delta uint64) {
	atomic.AddUint64(&s.LoadErrorCount, delta)
}

// AddEvictionCount increases EvictionCount atomically.
func (s *Stats) AddEvictionCount(delta uint64) {
	atomic.AddUint64(&s.EvictionCount, delta)
}

// Copy copies current stats to t.
func (s *Stats) Copy(t *Stats) {
	t.HitCount = atomic.LoadUint64(&s.HitCount)
	t.MissCount = atomic.LoadUint64(&s.MissCount)
	t.LoadSuccessCount = atomic.LoadUint64(&s.LoadSuccessCount)
	t.LoadErrorCount = atomic.LoadUint64(&s.LoadErrorCount)
	t.EvictionCount = atomic.LoadUint64(&s.EvictionCount)
}
