package report

import (
	"os"
	"testing"
)

func TestRequestWikipedia(t *testing.T) {
	for _, p := range policies {
		testRequestWikipedia(t, p, "request_wikipedia-"+p+".txt")
	}
}

func testRequestWikipedia(t *testing.T, policy, reportFile string) {
	traceFiles := "wiki.[1-9]*"
	opt := options{
		policy:         policy,
		cacheSize:      512,
		reportInterval: 10000,
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

func TestSizeWikipedia(t *testing.T) {
	for _, p := range policies {
		testSizeWikipedia(t, p, "size_wikipedia-"+p+".txt")
	}
}

func testSizeWikipedia(t *testing.T, policy, reportFile string) {
	traceFiles := "wiki.[1-9]*"
	opt := options{
		policy:   policy,
		maxItems: 100000,
	}

	r, err := openFilesGlob(traceFiles)
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()
	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)

	for i := 500; i <= 5000; i += 500 {
		opt.cacheSize = i
		provider := NewWikipediaProvider(r)
		benchmarkCache(provider, reporter, opt)
		err = r.Reset()
		if err != nil {
			t.Fatal(err)
		}
	}
}
