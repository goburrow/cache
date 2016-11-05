package report

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
)

type umassProvider struct {
	r *bufio.Reader
}

func NewUMassProvider(r io.Reader) Provider {
	return &umassProvider{
		r: bufio.NewReader(r),
	}
}

func (p *umassProvider) Provide(keys chan<- interface{}, done <-chan struct{}) {
	defer close(keys)
	for {
		b, err := p.r.ReadBytes('\n')
		if err != nil {
			return
		}
		cols := bytes.Split(b, []byte{','})
		if !p.valid(cols) {
			continue
		}
		start, end, err := p.parse(cols)
		if err != nil {
			return
		}
		for i := start; i < end; i++ {
			select {
			case <-done:
				return
			case keys <- i:
			}
		}
	}
}

func (p *umassProvider) valid(cols [][]byte) bool {
	if len(cols) < 4 {
		return false
	}
	for _, c := range cols {
		if len(c) == 0 {
			return false
		}
	}
	opcode := cols[3][0]
	if opcode != 'R' && opcode != 'r' {
		return false
	}
	return true
}

func (p *umassProvider) parse(cols [][]byte) (start, end uint64, err error) {
	const blockSize = 512

	start, err = strconv.ParseUint(string(cols[1]), 10, 64)
	if err != nil {
		return
	}
	size, err := strconv.ParseUint(string(cols[2]), 10, 64)
	if err != nil {
		return
	}
	end = start + (size+(blockSize-1))/blockSize
	return
}
