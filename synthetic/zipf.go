package synthetic

import (
	"math/rand"
	"time"
)

type zipfGenerator struct {
	r   *rand.Zipf
	min int
}

func (g *zipfGenerator) Int() int {
	return g.min + int(g.r.Uint64())
}

// Zipf returns a Generator resembling Zipf distribution.
func Zipf(min, max int, exp float64) Generator {
	if max <= min {
		panic("synthetic: invalid zipf range")
	}
	if exp <= 1.0 {
		panic("synthetic: invalid zipf exponent")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &zipfGenerator{
		r:   rand.NewZipf(r, exp, 1.0, uint64(max-min)),
		min: min,
	}
}
