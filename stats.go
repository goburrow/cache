package cache

import (
	"sync/atomic"
	"time"
)

// Stats is statistics about performance of a cache.
type Stats struct {
	HitCount         uint64
	MissCount        uint64
	LoadSuccessCount uint64
	LoadErrorCount   uint64
	TotalLoadTime    time.Duration
	EvictionCount    uint64
}

// StatsCounter accumulates statistics of a cache.
type StatsCounter interface {
	// RecordHits records cache hits.
	RecordHits(count uint64)

	// RecordMisses records cache misses.
	RecordMisses(count uint64)

	// RecordLoadSuccess records successful load of a new entry.
	RecordLoadSuccess(loadTime time.Duration)

	// RecordLoadError records failed load of a new entry.
	RecordLoadError(loadTime time.Duration)

	// RecordEviction records eviction of an entry from the cache.
	RecordEviction()

	// Snapshot writes snapshot of this counter values to the given Stats pointer.
	Snapshot(*Stats)
}

// statsCounter is a simple implementation of StatsCounter.
type statsCounter struct {
	Stats
}

// RecordHits increases HitCount atomically.
func (s *statsCounter) RecordHits(count uint64) {
	atomic.AddUint64(&s.Stats.HitCount, count)
}

// RecordMisses increases MissCount atomically.
func (s *statsCounter) RecordMisses(count uint64) {
	atomic.AddUint64(&s.Stats.MissCount, count)
}

// RecordLoadSuccess increases LoadSuccessCount atomically.
func (s *statsCounter) RecordLoadSuccess(loadTime time.Duration) {
	atomic.AddUint64(&s.Stats.LoadSuccessCount, 1)
	atomic.AddInt64((*int64)(&s.Stats.TotalLoadTime), int64(loadTime))
}

// RecordLoadError increases LoadErrorCount atomically.
func (s *statsCounter) RecordLoadError(loadTime time.Duration) {
	atomic.AddUint64(&s.Stats.LoadErrorCount, 1)
	atomic.AddInt64((*int64)(&s.Stats.TotalLoadTime), int64(loadTime))
}

// RecordEviction increases EvictionCount atomically.
func (s *statsCounter) RecordEviction() {
	atomic.AddUint64(&s.Stats.EvictionCount, 1)
}

// Snapshot copies current stats to t.
func (s *statsCounter) Snapshot(t *Stats) {
	t.HitCount = atomic.LoadUint64(&s.HitCount)
	t.MissCount = atomic.LoadUint64(&s.MissCount)
	t.LoadSuccessCount = atomic.LoadUint64(&s.LoadSuccessCount)
	t.LoadErrorCount = atomic.LoadUint64(&s.LoadErrorCount)
	t.TotalLoadTime = time.Duration(atomic.LoadInt64((*int64)(&s.TotalLoadTime)))
	t.EvictionCount = atomic.LoadUint64(&s.EvictionCount)
}
