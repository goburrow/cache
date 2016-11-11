package traces

import "testing"

func TestRequestYouTube(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:         p,
				cacheSize:      512,
				reportInterval: 2000,
				maxItems:       200000,
			}
			testRequest(t, NewYoutubeProvider, opt,
				"youtube.parsed.0803*.dat", "request_youtube-"+p+".txt")
		})
	}
}

func TestSizeYouTube(t *testing.T) {
	for _, p := range policies {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			opt := options{
				policy:    p,
				cacheSize: 250,
				maxItems:  100000,
			}
			testSize(t, NewYoutubeProvider, opt,
				"youtube.parsed.0803*.dat", "size_youtube-"+p+".txt")
		})
	}
}
