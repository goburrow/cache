package traces

import "testing"

func TestRequestORMBusy(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      512,
				reportInterval: 40000,
				maxItems:       4000000,
			}
			testRequest(t, NewCache2kProvider, opt,
				"trace-mt-db-*-busy.trc.bin.bz2", "request_ormbusy-"+p+".txt")
		})
	}
}

func TestSizeORMBusy(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:    p,
				cacheSize: 250,
				maxItems:  1000000,
			}
			testSize(t, NewCache2kProvider, opt,
				"trace-mt-db-*-busy.trc.bin.bz2", "size_ormbusy-"+p+".txt")
		})
	}
}

func TestRequestCPP(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      128,
				reportInterval: 100,
				maxItems:       9000,
			}
			testRequest(t, NewCache2kProvider, opt,
				"trace-cpp.trc.bin.gz", "request_cpp-"+p+".txt")
		})
	}
}

func TestSizeCPP(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:    p,
				cacheSize: 25,
				maxItems:  9000,
			}
			testSize(t, NewCache2kProvider, opt,
				"trace-cpp.trc.bin.gz", "size_cpp-"+p+".txt")
		})
	}
}

func TestRequestGlimpse(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      512,
				reportInterval: 100,
				maxItems:       6000,
			}
			testRequest(t, NewCache2kProvider, opt,
				"trace-glimpse.trc.bin.gz", "request_glimpse-"+p+".txt")
		})
	}
}

func TestSizeGlimpse(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:    p,
				cacheSize: 125,
				maxItems:  6000,
			}
			testSize(t, NewCache2kProvider, opt,
				"trace-glimpse.trc.bin.gz", "size_glimpse-"+p+".txt")
		})
	}
}

func TestRequestOLTP(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      512,
				reportInterval: 1000,
				maxItems:       900000,
			}
			testRequest(t, NewCache2kProvider, opt,
				"trace-oltp.trc.bin.gz", "request_oltp-"+p+".txt")
		})
	}
}

func TestSizeOLTP(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:    p,
				cacheSize: 250,
				maxItems:  500000,
			}
			testSize(t, NewCache2kProvider, opt,
				"trace-oltp.trc.bin.gz", "size_oltp-"+p+".txt")
		})
	}
}

func TestRequestSprite(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      512,
				reportInterval: 1000,
				maxItems:       120000,
			}
			testRequest(t, NewCache2kProvider, opt,
				"trace-sprite.trc.bin.gz", "request_sprite-"+p+".txt")
		})
	}
}

func TestSizeSprite(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:    p,
				cacheSize: 25,
				maxItems:  120000,
			}
			testSize(t, NewCache2kProvider, opt,
				"trace-sprite.trc.bin.gz", "size_sprite-"+p+".txt")
		})
	}
}

func TestRequestMulti2(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      512,
				reportInterval: 200,
				maxItems:       25000,
			}
			testRequest(t, NewCache2kProvider, opt,
				"trace-multi2.trc.bin.gz", "request_multi2-"+p+".txt")
		})
	}
}

func TestSizeMulti2(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:    p,
				cacheSize: 250,
				maxItems:  25000,
			}
			testSize(t, NewCache2kProvider, opt,
				"trace-multi2.trc.bin.gz", "size_multi2-"+p+".txt")
		})
	}
}
