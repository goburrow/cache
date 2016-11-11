package traces

import "testing"

func TestRequestWikipedia(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      512,
				reportInterval: 10000,
				maxItems:       1000000,
			}
			testRequest(t, NewWikipediaProvider, opt,
				"wiki.*.gz", "request_wikipedia-"+p+".txt")
		})
	}
}

func TestSizeWikipedia(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:    p,
				cacheSize: 250,
				maxItems:  100000,
			}
			testSize(t, NewWikipediaProvider, opt,
				"wiki.*.gz", "size_wikipedia-"+p+".txt")
		})
	}
}
