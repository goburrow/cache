package synthetic

import (
	"math/rand"
	"time"
)

type hotspotGenerator struct {
	r   *rand.Rand
	min int

	hotFrac float64
	hotN    int
	coldN   int
}

func (g *hotspotGenerator) Int() int {
	v := g.min
	r := g.r.Float64()
	if r > g.hotFrac {
		// Hotset
		v += g.r.Intn(g.hotN)
	} else {
		// Coldset
		v += g.hotN + g.r.Intn(g.coldN)
	}
	return v
}

// Hotspot returns a Generator resembling a hotspot distribution.
// hotFrac is the fraction of total items which have a proportion (1.0-hotFrac).
func Hotspot(min, max int, hotFrac float64) Generator {
	if max <= min {
		panic("synthetic: invalid hotspot range")
	}
	if hotFrac < 0.0 || hotFrac > 1.0 {
		panic("synthetic: invalid hotspot fraction")
	}
	n := max - min
	hotN := int(hotFrac * float64(n))
	coldN := n - hotN

	return &hotspotGenerator{
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		min:     min,
		hotFrac: hotFrac,
		hotN:    hotN,
		coldN:   coldN,
	}
}
