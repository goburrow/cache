package report

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goburrow/cache"
)

type Reporter interface {
	Report(cache.Stats, options)
}

type Provider interface {
	Provide(keys chan<- interface{}, done <-chan struct{})
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
	done := make(chan struct{})
	defer close(done)

	go p.Provide(keys, done)
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

type filesReader struct {
	io.Reader
	files []*os.File
}

func openFilesGlob(pattern string) (*filesReader, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("%s not found", pattern)
	}
	return openFiles(files...)
}

func openFiles(files ...string) (*filesReader, error) {
	r := &filesReader{}
	r.files = make([]*os.File, 0, len(files))
	readers := make([]io.Reader, 0, len(files))
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			r.Close()
			return nil, err
		}
		r.files = append(r.files, f)
		readers = append(readers, f)
	}
	r.Reader = io.MultiReader(readers...)
	return r, nil
}

func (r *filesReader) Close() error {
	var err error
	for _, f := range r.files {
		e := f.Close()
		if err != nil && e != nil {
			err = e
		}
	}
	return err
}

func (r *filesReader) Reset() error {
	readers := make([]io.Reader, 0, len(r.files))
	for _, f := range r.files {
		_, err := f.Seek(0, 0)
		if err != nil {
			return err
		}
		readers = append(readers, f)
	}
	r.Reader = io.MultiReader(readers...)
	return nil
}
