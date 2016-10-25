package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/goburrow/cache"
)

func load(k cache.Key) (cache.Value, error) {
	return fmt.Sprintf("%d", k), nil
}

func report(c cache.Cache) {
	var st cache.Stats
	c.Stats(&st)

	total := st.HitCount + st.MissCount
	if total == 0 {
		return
	}
	hitPerc := float64(st.HitCount) / float64(total) * 100.0
	missPerc := float64(st.MissCount) / float64(total) * 100.0

	fmt.Printf("total: %d, hit: %d (%.2f%%), miss: %d (%.2f%%), eviction: %d, load: %s\n",
		total, st.HitCount, hitPerc, st.MissCount, missPerc, st.EvictionCount, st.TotalLoadTime)
}

func main() {
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
			report(c)
		}
	}
}
