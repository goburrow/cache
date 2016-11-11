package traces

import (
	"context"
	"fmt"
	"io"

	"github.com/goburrow/cache"
)

type Reporter interface {
	Report(cache.Stats, options)
}

type Provider interface {
	Provide(ctx context.Context, keys chan<- interface{})
}

type reporter struct {
	w             io.Writer
	headerPrinted bool
}

func NewReporter(w io.Writer) Reporter {
	return &reporter{w: w}
}

func (r *reporter) Report(st cache.Stats, opt options) {
	if !r.headerPrinted {
		fmt.Fprintf(r.w, "Requets,Hits,HitRate,Evictions,CacheSize\n")
		r.headerPrinted = true
	}
	fmt.Fprintf(r.w, "%d,%d,%.04f,%d,%d\n",
		st.RequestCount(), st.HitCount, st.HitRate(), st.EvictionCount,
		opt.cacheSize)
}

type options struct {
	policy         string
	cacheSize      int
	reportInterval int
	maxItems       int
}

var policies = []string{
	"lru",
	"slru",
	"tinylfu",
}

func benchmarkCache(p Provider, r Reporter, opt options) {
	c := cache.New(cache.WithMaximumSize(opt.cacheSize), cache.WithPolicy(opt.policy))
	defer c.Close()

	keys := make(chan interface{}, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go p.Provide(ctx, keys)
	stats := cache.Stats{}
	i := 0
	for {
		if opt.maxItems > 0 && i >= opt.maxItems {
			break
		}
		k, ok := <-keys
		if !ok {
			break
		}
		_, ok = c.GetIfPresent(k)
		if !ok {
			c.Put(k, k)
		}
		i++
		if opt.reportInterval > 0 && i%opt.reportInterval == 0 {
			c.Stats(&stats)
			r.Report(stats, opt)
		}
	}
	if opt.reportInterval == 0 {
		c.Stats(&stats)
		r.Report(stats, opt)
	}
}
