package cache

import (
	"testing"
	"time"

	"github.com/goburrow/cache/synthetic"
)

const (
	testMaxSize        = 512
	benchmarkThreshold = 100
)

type sameGenerator int

func (g sameGenerator) Int() int {
	return int(g)
}

func BenchmarkSame(b *testing.B) {
	g := sameGenerator(1)
	benchmarkCache(b, g)
}

func BenchmarkUniform(b *testing.B) {
	distintKeys := testMaxSize * 2
	g := synthetic.Uniform(0, distintKeys)
	benchmarkCache(b, g)
}

func BenchmarkUniformLess(b *testing.B) {
	distintKeys := testMaxSize
	g := synthetic.Uniform(0, distintKeys)
	benchmarkCache(b, g)
}

func BenchmarkCounter(b *testing.B) {
	g := synthetic.Counter(0)
	benchmarkCache(b, g)
}

func BenchmarkExponential(b *testing.B) {
	g := synthetic.Exponential(1.0)
	benchmarkCache(b, g)
}

func BenchmarkZipf(b *testing.B) {
	items := testMaxSize * 10
	g := synthetic.Zipf(0, items, 1.01)
	benchmarkCache(b, g)
}

func BenchmarkHotspot(b *testing.B) {
	items := testMaxSize * 2
	g := synthetic.Hotspot(0, items, 0.25)
	benchmarkCache(b, g)
}

func benchmarkCache(b *testing.B, g synthetic.Generator) {
	c := New(WithMaximumSize(testMaxSize))
	defer c.Close()

	intCh := make(chan int, 100)
	go func(n int) {
		for i := 0; i < n; i++ {
			intCh <- g.Int()
		}
	}(b.N)
	defer close(intCh)

	if b.N > benchmarkThreshold {
		defer printStats(b, c, time.Now())
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := <-intCh
			_, ok := c.GetIfPresent(k)
			if !ok {
				c.Put(k, k)
			}
		}
	})
}

func printStats(b *testing.B, c Cache, start time.Time) {
	dur := time.Since(start)

	var st Stats
	c.Stats(&st)

	b.Logf("total: %d (%s), hits: %d (%.2f%%), misses: %d (%.2f%%), evictions: %d\n",
		st.RequestCount(), dur,
		st.HitCount, st.HitRate()*100.0,
		st.MissCount, st.MissRate()*100.0,
		st.EvictionCount)
}
