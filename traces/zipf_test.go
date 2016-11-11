package traces

import (
	"os"
	"testing"
)

func TestRequestZipf(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			testRequestZipf(t, p, "request_zipf-"+p+".txt")
		})
	}
}

func testRequestZipf(t *testing.T, policy, reportFile string) {
	opt := options{
		policy:         policy,
		cacheSize:      512,
		reportInterval: 1000,
		maxItems:       100000,
	}

	provider := NewZipfProvider(1.01, opt.maxItems)

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	benchmarkCache(provider, reporter, opt)
}

func TestSizeZipf(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			testSizeZipf(t, p, "size_zipf-"+p+".txt")
		})
	}
}

func testSizeZipf(t *testing.T, policy, reportFile string) {
	opt := options{
		cacheSize: 250,
		policy:    policy,
		maxItems:  100000,
	}

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	for i := 0; i < 5; i++ {
		provider := NewZipfProvider(1.01, opt.maxItems)
		benchmarkCache(provider, reporter, opt)
		opt.cacheSize += opt.cacheSize
	}
}
