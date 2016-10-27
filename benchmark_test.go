package cache

import (
	"math/rand"
	"testing"
	"time"
)

const (
	distinctKeys    = 4096
	reportThreshold = 10000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func BenchmarkCache(b *testing.B) {
	c := New(WithMaximumSize(distinctKeys))
	benchmarkCache(b, c)
}

func BenchmarkCacheHalf(b *testing.B) {
	c := New(WithMaximumSize(distinctKeys / 2))
	benchmarkCache(b, c)
}

func benchmarkCache(b *testing.B, c Cache) {
	defer c.Close()
	b.ResetTimer()
	b.ReportAllocs()

	if b.N > reportThreshold {
		defer printStats(b, c, time.Now())
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := rand.Intn(distinctKeys)
			c.Put(k, b)
			k = rand.Intn(distinctKeys)
			_, _ = c.GetIfPresent(k)
		}
	})
}

func printStats(b *testing.B, c Cache, start time.Time) {
	dur := time.Since(start)

	var st Stats
	c.Stats(&st)

	b.Logf("total: %d (%s), hits: %d (%.2f%%), misses: %d (%.2f%%), evictions: %d, load: %s (%s)\n",
		st.RequestCount(), dur, st.HitCount, st.HitRate()*100.0, st.MissCount, st.MissRate()*100.0,
		st.EvictionCount, st.TotalLoadTime, st.AverageLoadPenalty())
}
