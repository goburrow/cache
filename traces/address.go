package traces

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strconv"
)

type addressProvider struct {
	r *bufio.Reader
}

// NewAddressProvider returns a Provider with items are from
// application traces by the University of California, San Diego
// (http://cseweb.ucsd.edu/classes/fa07/cse240a/project1.html).
func NewAddressProvider(r io.Reader) Provider {
	return &addressProvider{
		r: bufio.NewReader(r),
	}
}

func (p *addressProvider) Provide(ctx context.Context, keys chan<- interface{}) {
	defer close(keys)
	for {
		b, err := p.r.ReadBytes('\n')
		if err != nil {
			return
		}
		v := p.parse(b)
		if v > 0 {
			select {
			case <-ctx.Done():
				return
			case keys <- v:
			}
		}
	}
}

func (p *addressProvider) parse(b []byte) uint64 {
	idx := bytes.IndexByte(b, ' ')
	if idx < 0 {
		return 0
	}
	b = b[idx+1:]
	idx = bytes.IndexByte(b, ' ')
	if idx < 0 {
		return 0
	}
	b = b[:idx]

	val, err := strconv.ParseUint(string(b), 0, 0)
	if err != nil {
		return 0
	}
	return val
}
