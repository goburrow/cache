package report

import (
	"os"
	"testing"
)

func TestWikipediaLRU(t *testing.T) {
	testWikipedia(t, "lru", "wikipedia-lru.txt")
}

func TestWikipediaSLRU(t *testing.T) {
	testWikipedia(t, "slru", "wikipedia-slru.txt")
}

func TestWikipediaTinyLFU(t *testing.T) {
	testWikipedia(t, "tinylfu", "wikipedia-tinylfu.txt")
}

func testWikipedia(t *testing.T, policy, reportFile string) {
	traceFiles := "wiki.[1-9]*"
	opt := options{
		policy:         policy,
		cacheSize:      512,
		reportInterval: 1000,
		maxItems:       1000000,
	}

	r, err := openFilesGlob(traceFiles)
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()
	provider := NewWikipediaProvider(r)

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	benchmarkCache(provider, reporter, opt)
}
