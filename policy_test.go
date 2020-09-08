package cache

import (
	"sync/atomic"
	"testing"
)

func BenchmarkCacheSegment(b *testing.B) {
	c := cache{}
	const count = 1 << 10
	entries := make([]*entry, count)
	for i := range entries {
		entries[i] = newEntry(i, i, uint64(i))
	}
	var n int32
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddInt32(&n, 1)
			c.getOrSet(entries[i&(count-1)])
			if i > 0 && i&0xf == 0 {
				c.delete(entries[(i-1)&(count-1)])
			}
		}
	})
}
