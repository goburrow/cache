package traces

import (
	"io"
	"os"
	"testing"
)

func testRequest(t *testing.T, newProvider func(io.Reader) Provider, opt options, traceFiles string, reportFile string) {
	r, err := openFilesGlob(traceFiles)
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()
	provider := newProvider(r)

	w, err := os.Create(reportFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	reporter := NewReporter(w)
	benchmarkCache(provider, reporter, opt)
}

func testSize(t *testing.T, newProvider func(io.Reader) Provider, opt options, traceFiles, reportFile string) {
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
	for i := 0; i < 5; i++ {
		provider := newProvider(r)
		benchmarkCache(provider, reporter, opt)
		err = r.Reset()
		if err != nil {
			t.Fatal(err)
		}
		opt.cacheSize += opt.cacheSize
	}
}
