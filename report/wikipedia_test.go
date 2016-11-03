package report

import (
	"os"
	"testing"
)

func TestWikipedia(t *testing.T) {
	const (
		traceFiles = "wiki.[1-9]*"
		reportFile = "report-wikipedia.txt"
	)
	opt := options{
		cacheSize:      512,
		reportInterval: 1000,
		maxItems:       5000000,
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
