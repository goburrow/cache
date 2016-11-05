package report

import (
	"os"
	"testing"
)

func TestUMass(t *testing.T) {
	const (
		traceFiles = "WebSearch*.spc"
		reportFile = "report-umass.txt"
	)
	opt := options{
		cacheSize:      1000,
		reportInterval: 1000,
		maxItems:       5000000,
	}
	r, err := openFilesGlob(traceFiles)
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()
	provider := NewUMassProvider(r)

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	benchmarkCache(provider, reporter, opt)
}
