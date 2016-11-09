package report

import (
	"os"
	"testing"
)

func TestRequestWebSearch(t *testing.T) {
	for _, p := range policies {
		testRequestWebSearch(t, p, "request_websearch-"+p+".txt")
	}
}

func testRequestWebSearch(t *testing.T, policy, reportFile string) {
	traceFiles := "WebSearch*.spc.bz2"
	opt := options{
		policy:         policy,
		cacheSize:      512000,
		reportInterval: 100000,
		maxItems:       10000000,
	}
	r, err := openFilesGlob(traceFiles)
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()
	provider := NewWebSearchProvider(r)

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	benchmarkCache(provider, reporter, opt)
}
