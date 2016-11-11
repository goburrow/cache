package traces

import "testing"

func TestRequestAddress(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      128,
				reportInterval: 5000,
				maxItems:       500000,
			}
			testRequest(t, NewAddressProvider, opt,
				"traces/gcc.trace", "request_address-"+p+".txt")
		})
	}
}

func TestSizeAddress(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:    p,
				cacheSize: 25,
				maxItems:  100000,
			}
			testSize(t, NewAddressProvider, opt,
				"traces/gcc.trace", "size_address-"+p+".txt")
		})
	}
}
