package report

import (
	"os"
	"testing"
)

func TestYouTubeLRU(t *testing.T) {
	testYouTube(t, "lru", "youtube-lru.txt")
}

func TestYouTubeSLRU(t *testing.T) {
	testYouTube(t, "slru", "youtube-slru.txt")
}

func TestYouTubeTinyLFU(t *testing.T) {
	testYouTube(t, "tinylfu", "youtube-tinylfu.txt")
}

func testYouTube(t *testing.T, policy, reportFile string) {
	traceFiles := "youtube.parsed.*.dat"
	opt := options{
		policy:         policy,
		cacheSize:      1024,
		reportInterval: 1000,
		maxItems:       2000000,
	}

	r, err := openFilesGlob(traceFiles)
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()
	provider := NewYoutubeProvider(r)

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	benchmarkCache(provider, reporter, opt)
}
