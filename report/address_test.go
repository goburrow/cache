package report

import (
	"os"
	"testing"
)

func TestRequestAddress(t *testing.T) {
	for _, p := range policies {
		testRequestAddress(t, p, "request_address-"+p+".txt")
	}
}

func testRequestAddress(t *testing.T, policy, reportFile string) {
	traceFiles := "traces/gcc.trace"
	opt := options{
		policy:         policy,
		cacheSize:      128,
		reportInterval: 5000,
		maxItems:       500000,
	}

	r, err := openFilesGlob(traceFiles)
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()
	provider := NewAddressProvider(r)

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	benchmarkCache(provider, reporter, opt)
}

func TestSizeAddress(t *testing.T) {
	for _, p := range policies {
		testSizeAddress(t, p, "size_address-"+p+".txt")
	}
}

func testSizeAddress(t *testing.T, policy, reportFile string) {
	traceFiles := "traces/gcc.trace"
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

	for i := 50; i <= 500; i += 50 {
		opt.cacheSize = i
		provider := NewAddressProvider(r)
		benchmarkCache(provider, reporter, opt)
		err = r.Reset()
		if err != nil {
			t.Fatal(err)
		}
	}
}
