package cache

import (
	"testing"
	"time"
)

func TestStatsCounter(t *testing.T) {
	c := statsCounter{}
	c.RecordHits(3)
	c.RecordMisses(2)
	c.RecordLoadSuccess(2 * time.Second)
	c.RecordLoadError(1 * time.Second)
	c.RecordEviction()

	var st Stats
	c.Snapshot(&st)

	if st.HitCount != 3 {
		t.Fatalf("unexpected hit count: %v", st)
	}
	if st.MissCount != 2 {
		t.Fatalf("unexpected miss count: %v", st)
	}
	if st.LoadSuccessCount != 1 {
		t.Fatalf("unexpected success count: %v", st)
	}
	if st.LoadErrorCount != 1 {
		t.Fatalf("unexpected error count: %v", st)
	}
	if st.TotalLoadTime != 3*time.Second {
		t.Fatalf("unexpected load time: %v", st)
	}
	if st.EvictionCount != 1 {
		t.Fatalf("unexpected eviction count: %v", st)
	}
}
