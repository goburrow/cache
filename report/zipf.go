package report

import "math/rand"

type zipfProvider struct {
	r *rand.Zipf
	n int
}

func NewZipfProvider(s float64, num int) Provider {
	if s <= 1.0 || num <= 0 {
		panic("invalid zipf parameters")
	}
	r := rand.New(rand.NewSource(1))
	return &zipfProvider{
		r: rand.NewZipf(r, s, 1.0, 1<<31-1),
		n: num,
	}
}

func (p *zipfProvider) Provide(keys chan<- interface{}, done <-chan struct{}) {
	defer close(keys)
	for i := 0; i < p.n; i++ {
		v := p.r.Uint64()
		select {
		case <-done:
			return
		case keys <- v:
		}
	}
}
