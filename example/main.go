package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/goburrow/cache"
)

func main() {
	load := func(k cache.Key) (cache.Value, error) {
		return fmt.Sprintf("%d", k), nil
	}
	// Create a new cache
	c := cache.NewLoadingCache(load,
		cache.WithMaximumSize(1000),
		cache.WithExpireAfterAccess(10*time.Second),
		cache.WithRefreshAfterWrite(60*time.Second),
	)

	getTicker := time.Tick(10 * time.Millisecond)
	reportTicker := time.Tick(1 * time.Second)
	for {
		select {
		case <-getTicker:
			_, _ = c.Get(rand.Intn(2000))
		case <-reportTicker:
			st := cache.Stats{}
			c.Stats(&st)
			fmt.Printf("total: %d, hits: %d (%.2f%%), misses: %d (%.2f%%), evictions: %d, load: %s (%s)\n",
				st.RequestCount(), st.HitCount, st.HitRate()*100.0, st.MissCount, st.MissRate()*100.0,
				st.EvictionCount, st.TotalLoadTime, st.AverageLoadPenalty())
		}
	}
}
