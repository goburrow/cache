package report

import (
	"bufio"
	"bytes"
	"io"
)

type wikipediaProvider struct {
	r *bufio.Reader
}

func NewWikipediaProvider(r io.Reader) Provider {
	return &wikipediaProvider{
		r: bufio.NewReader(r),
	}
}

func (p *wikipediaProvider) Provide(keys chan<- interface{}, done <-chan struct{}) {
	defer close(keys)
	for {
		b, err := p.r.ReadBytes('\n')
		if err != nil {
			return
		}
		v := p.parse(b)
		if v != "" {
			select {
			case <-done:
				return
			case keys <- v:
			}
		}
	}
}

func (p *wikipediaProvider) parse(b []byte) string {
	// Get url
	idx := bytes.Index(b, []byte("http://"))
	if idx < 0 {
		return ""
	}
	b = b[idx+len("http://"):]
	// Get path
	idx = bytes.IndexByte(b, '/')
	if idx > 0 {
		b = b[idx:]
	}
	// Skip params
	idx = bytes.IndexAny(b, "? ")
	if idx > 0 {
		b = b[:idx]
	}
	return string(b)
}
