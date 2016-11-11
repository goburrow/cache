package traces

import (
	"bufio"
	"bytes"
	"context"
	"io"
)

type youtubeProvider struct {
	r *bufio.Reader
}

func NewYoutubeProvider(r io.Reader) Provider {
	return &youtubeProvider{
		r: bufio.NewReader(r),
	}
}

func (p *youtubeProvider) Provide(ctx context.Context, keys chan<- interface{}) {
	defer close(keys)
	for {
		b, err := p.r.ReadBytes('\n')
		if err != nil {
			return
		}
		v := p.parse(b)
		if v != "" {
			select {
			case <-ctx.Done():
				return
			case keys <- v:
			}
		}
	}
}

func (p *youtubeProvider) parse(b []byte) string {
	// Get video id
	idx := bytes.Index(b, []byte("GETVIDEO "))
	if idx < 0 {
		return ""
	}
	b = b[idx+len("GETVIDEO "):]
	idx = bytes.IndexAny(b, "& ")
	if idx > 0 {
		b = b[:idx]
	}
	return string(b)
}
