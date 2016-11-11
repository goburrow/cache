package traces

import "testing"

func TestRequestWebSearch(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      256000,
				reportInterval: 10000,
				maxItems:       1000000,
			}
			testRequest(t, NewStorageProvider, opt,
				"WebSearch*.spc.bz2", "request_websearch-"+p+".txt")
		})
	}
}

func TestRequestFinancial(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      512,
				reportInterval: 30000,
				maxItems:       3000000,
			}
			testRequest(t, NewStorageProvider, opt,
				"Financial*.spc.bz2", "request_financial-"+p+".txt")
		})
	}
}

func TestSizeWebSearch(t *testing.T) {
	for _, p := range policies {
		p := p
		opt := options{
			policy:    p,
			cacheSize: 25000,
			maxItems:  1000000,
		}
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			testSize(t, NewStorageProvider, opt,
				"WebSearch*.spc.bz2", "size_websearch-"+p+".txt")
		})
	}
}

func TestSizeFinancial(t *testing.T) {
	for _, p := range policies {
		p := p
		opt := options{
			policy:    p,
			cacheSize: 250,
			maxItems:  1000000,
		}
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			testSize(t, NewStorageProvider, opt,
				"Financial*.spc.bz2", "size_financial-"+p+".txt")
		})
	}
}
