package report

import (
	"os"
	"testing"
)

func TestRequestYouTube(t *testing.T) {
	for _, p := range policies {
		testRequestYouTube(t, p, "request_youtube-"+p+".txt")
	}
}

func testRequestYouTube(t *testing.T, policy, reportFile string) {
	traceFiles := "youtube.parsed.0803*.dat"
	opt := options{
		policy:         policy,
		cacheSize:      512,
		reportInterval: 2000,
		maxItems:       200000,
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

func TestSizeYouTube(t *testing.T) {
	for _, p := range policies {
		testSizeYouTube(t, p, "size_youtube-"+p+".txt")
	}
}

func testSizeYouTube(t *testing.T, policy, reportFile string) {
	traceFiles := "youtube.parsed.0803*.dat"
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
		provider := NewYoutubeProvider(r)
		benchmarkCache(provider, reporter, opt)
		err = r.Reset()
		if err != nil {
			t.Fatal(err)
		}
	}
}
