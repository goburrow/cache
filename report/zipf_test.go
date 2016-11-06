package report

import (
	"os"
	"testing"
)

func TestZipfLRU(t *testing.T) {
	testZipf(t, "lru", "zipf-lru.txt")
}

func TestZipfSLRU(t *testing.T) {
	testZipf(t, "slru", "zipf-slru.txt")
}

func TestZipfTinyLFU(t *testing.T) {
	testZipf(t, "tinylfu", "zipf-tinylfu.txt")
}

func testZipf(t *testing.T, policy, reportFile string) {
	opt := options{
		policy:         policy,
		cacheSize:      512,
		reportInterval: 1000,
		maxItems:       1000000,
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
