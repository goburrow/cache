package report

import (
	"os"
	"testing"
)

func TestYoutube(t *testing.T) {
	const (
		traceFiles = "youtube.parsed.*.dat"
		reportFile = "report-youtube.txt"
	)
	opt := options{
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
