package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/goburrow/cache"
)

func main() {
	load := func(k cache.Key) (cache.Value, error) {
		fmt.Printf("loading %v\n", k)
		time.Sleep(500 * time.Millisecond)
		return fmt.Sprintf("%d-%d", k, time.Now().Unix()), nil
	}
	remove := func(k cache.Key, v cache.Value) {
		fmt.Printf("removed %v (%v)\n", k, v)
	}
	// Create a new cache
	c := cache.NewLoadingCache(load,
		cache.WithMaximumSize(1000),
		cache.WithExpireAfterAccess(30*time.Second),
		cache.WithRefreshAfterWrite(20*time.Second),
		cache.WithRemovalListener(remove),
	)

	getTicker := time.Tick(2 * time.Second)
	reportTicker := time.Tick(30 * time.Second)
	for {
		select {
		case <-getTicker:
			k := rand.Intn(100)
			v, _ := c.Get(k)
			fmt.Printf("get %v: %v\n", k, v)
		case <-reportTicker:
			st := cache.Stats{}
			c.Stats(&st)
			fmt.Printf("total: %d, hits: %d (%.2f%%), misses: %d (%.2f%%), evictions: %d, load: %s (%s)\n",
				st.RequestCount(), st.HitCount, st.HitRate()*100.0, st.MissCount, st.MissRate()*100.0,
				st.EvictionCount, st.TotalLoadTime, st.AverageLoadPenalty())
		}
	}
}
