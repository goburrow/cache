package cache

const sketchDepth = 4

// countMinSketch is an implementation of count-min sketch with 4-bit counters.
// See http://dimacs.rutgers.edu/~graham/pubs/papers/cmsoft.pdf
type countMinSketch struct {
	counters [sketchDepth]nybbles
	mask     uint32
}

// init initialize count-min sketch with the given width.
func (c *countMinSketch) init(width int) {
	var size uint32
	if width > 0 {
		size = nextPowerOfTwo(uint32(width))
	} else {
		size = 1
	}
	for i := range c.counters {
		if len(c.counters[i]) != int(size/2) {
			c.counters[i] = newNybbles(size)
		} else {
			c.counters[i].clear()
		}
	}
	c.mask = size - 1
}

// add increases counters associated with the given hash.
func (c *countMinSketch) add(h uint64) {
	h1, h2 := uint32(h), uint32(h>>32)

	for i, row := range c.counters {
		pos := (h1 + uint32(i)*h2) & c.mask
		row.inc(pos)
	}
}

// estimate returns minimum value of counters associated with the given hash.
func (c *countMinSketch) estimate(h uint64) uint8 {
	h1, h2 := uint32(h), uint32(h>>32)

	var min uint8 = 0xFF
	for i, row := range c.counters {
		pos := (h1 + uint32(i)*h2) & c.mask
		v := row.get(pos)
		if v < min {
			min = v
		}
	}
	return min
}

// reset resets all counters.
func (c *countMinSketch) reset() {
	for _, row := range c.counters {
		row.reset()
	}
}

// nubbles is a nybble vector.
type nybbles []byte

func newNybbles(width uint32) nybbles {
	return make(nybbles, width/2)
}

// get returns value at index i.
func (n nybbles) get(i uint32) byte {
	idx := i / 2
	shift := (i & 1) * 4
	return byte(n[idx]>>shift) & 0x0f
}

// inc increases value at index i.
func (n nybbles) inc(i uint32) {
	idx := i / 2
	shift := (i & 1) * 4
	v := (n[idx] >> shift) & 0x0f
	if v < 15 {
		n[idx] += 1 << shift
	}
}

// reset divides all counters by two.
func (n nybbles) reset() {
	for i := range n {
		n[i] = (n[i] >> 1) & 0x77
	}
}

func (n nybbles) clear() {
	for i := range n {
		n[i] = 0
	}
}

// nextPowerOfTwo returns the smallest power of two which is greater than or equal to i.
func nextPowerOfTwo(i uint32) uint32 {
	n := i - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
