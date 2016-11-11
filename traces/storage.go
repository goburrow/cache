package traces

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strconv"
)

type storageProvider struct {
	r *bufio.Reader
}

// NewStorageProvider returns a Provider with items are from
// Storage traces by the University of Massachusetts
// (http://traces.cs.umass.edu/index.php/Storage/Storage).
func NewStorageProvider(r io.Reader) Provider {
	return &storageProvider{
		r: bufio.NewReader(r),
	}
}

func (p *storageProvider) Provide(ctx context.Context, keys chan<- interface{}) {
	defer close(keys)
	for {
		b, err := p.r.ReadBytes('\n')
		if err != nil {
			return
		}
		k := p.parse(b)
		if k > 0 {
			select {
			case <-ctx.Done():
				return
			case keys <- k:
			}
		}
	}
}

func (p *storageProvider) parse(b []byte) uint64 {
	idx := bytes.IndexByte(b, ',')
	if idx < 0 {
		return 0
	}
	b = b[idx+1:]
	idx = bytes.IndexByte(b, ',')
	if idx < 0 {
		return 0
	}
	k, err := strconv.ParseUint(string(b[:idx]), 10, 64)
	if err != nil {
		return 0
	}
	return k
}
