package synthetic

import (
	"math/rand"
	"time"
)

type uniformGenerator struct {
	r   *rand.Rand
	n   int
	min int
}

func (g *uniformGenerator) Int() int {
	return g.min + g.r.Intn(g.n)
}

// Uniform returns a Generator resembling a uniform distribution.
func Uniform(min, max int) Generator {
	if max <= min {
		panic("synthetic: invalid uniform range")
	}
	return &uniformGenerator{
		r:   rand.New(rand.NewSource(time.Now().UnixNano())),
		min: min,
		n:   max - min,
	}
}
