package report

import (
	"os"
	"testing"
)

func TestZipf(t *testing.T) {
	const (
		reportFile = "report-zipf.txt"
	)
	opt := options{
		cacheSize:      512,
		reportInterval: 1000,
		maxItems:       2000000,
	}

	provider := NewZipfProvider(1.1, opt.maxItems)

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	benchmarkCache(provider, reporter, opt)
}
