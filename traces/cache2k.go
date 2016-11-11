package traces

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
)

type cache2kProvider struct {
	r *bufio.Reader
}

// NewCache2kProvider returns a Provider which items are from traces
// in Cache2k repository (https://github.com/cache2k/cache2k-benchmark).
func NewCache2kProvider(r io.Reader) Provider {
	return &cache2kProvider{
		r: bufio.NewReader(r),
	}
}

func (p *cache2kProvider) Provide(ctx context.Context, keys chan<- interface{}) {
	defer close(keys)

	v := make([]byte, 4)
	for {
		_, err := p.r.Read(v)
		if err != nil {
			return
		}
		k := binary.LittleEndian.Uint32(v)
		select {
		case <-ctx.Done():
			return
		case keys <- k:
		}
	}
}
