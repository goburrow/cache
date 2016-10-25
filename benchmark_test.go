package cache

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkCache(b *testing.B) {
	maxSize := 4096
	distinctKeys := 4096
	benchmarkCache(b, maxSize, distinctKeys)
}

func BenchmarkCacheHalf(b *testing.B) {
	maxSize := 2048
	distinctKeys := 4096
	benchmarkCache(b, maxSize, distinctKeys)
}

func benchmarkCache(b *testing.B, maxSize, distinctKeys int) {
	c := New(WithMaximumSize(maxSize))
	defer c.Close()
	rand.Seed(time.Now().UnixNano())

	b.ResetTimer()
	b.ReportAllocs()

	if b.N > 100 {
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

	total := st.HitCount + st.MissCount
	hitPerc := float64(st.HitCount) / float64(total) * 100.0
	missPerc := float64(st.MissCount) / float64(total) * 100.0

	b.Logf("total: %d (%v), hit: %d (%.2f%%), miss: %d (%.2f%%), eviction: %d",
		total, dur, st.HitCount, hitPerc, st.MissCount, missPerc, st.EvictionCount)
}
