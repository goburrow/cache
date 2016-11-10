package report

import (
	"os"
	"testing"
)

func TestRequestWebSearch(t *testing.T) {
	opt := options{
		cacheSize:      256000,
		reportInterval: 10000,
		maxItems:       1000000,
	}
	for _, p := range policies {
		opt.policy = p
		testRequestStorage(t, opt, "WebSearch*.spc.bz2", "request_websearch-"+p+".txt")
	}
}

func TestRequestFinancial(t *testing.T) {
	opt := options{
		cacheSize:      512,
		reportInterval: 30000,
		maxItems:       3000000,
	}
	for _, p := range policies {
		opt.policy = p
		testRequestStorage(t, opt, "Financial*.spc.bz2", "request_financial-"+p+".txt")
	}
}

func testRequestStorage(t *testing.T, opt options, traceFiles string, reportFile string) {
	r, err := openFilesGlob(traceFiles)
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()
	provider := NewStorageProvider(r)

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	benchmarkCache(provider, reporter, opt)
}

func TestSizeWebSearch(t *testing.T) {
	opt := options{
		cacheSize: 50000,
		maxItems:  1000000,
	}
	for _, p := range policies {
		opt.policy = p
		testSizeStorage(t, opt, "WebSearch*.spc.bz2", "size_websearch-"+p+".txt")
	}
}

func TestSizeFinancial(t *testing.T) {
	opt := options{
		cacheSize: 500,
		maxItems:  1000000,
	}
	for _, p := range policies {
		opt.policy = p
		testSizeStorage(t, opt, "Financial*.spc.bz2", "size_financial-"+p+".txt")
	}
}

func testSizeStorage(t *testing.T, opt options, traceFiles, reportFile string) {
	cacheSize := opt.cacheSize
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
	for i := cacheSize; i < 5*cacheSize; i += cacheSize {
		opt.cacheSize = i
		provider := NewStorageProvider(r)
		benchmarkCache(provider, reporter, opt)
		err = r.Reset()
		if err != nil {
			t.Fatal(err)
		}
	}
}
