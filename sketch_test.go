package cache

import "testing"

func TestNybbles(t *testing.T) {
	var size uint32 = 8
	var i uint32
	n := newNybbles(size)
	for i = 0; i < size; i++ {
		// Increase value at i j times
		for j := i; j > 0; j-- {
			n.inc(i)
		}
	}
	for i = 0; i < size; i++ {
		assertNybble(t, n, i, byte(i))
	}
	for i = 0; i < 100; i++ {
		n.inc(1)
	}
	assertNybble(t, n, 1, 15)
	assertNybble(t, n, 0, 0)

	n.reset()
	assertNybble(t, n, 1, 7)
	assertNybble(t, n, 6, 3)

	n.reset()
	assertNybble(t, n, 6, 1)
	assertNybble(t, n, 2, 0)
}

func assertNybble(t *testing.T, n nybbles, idx uint32, expect byte) {
	v := n.get(idx)
	if v != expect {
		t.Fatalf("unexpected get(%d): 0x%x, want: %d", idx, v, expect)
	}
}

func TestCountMinSketch(t *testing.T) {
	cm := &countMinSketch{}
	cm.init(32)

	h := uint64(0xdeadbeef)
	cm.add(h)
	cm.add(h)

	n := cm.estimate(h)
	if n != 2 {
		t.Errorf("unexpected estimate: %d, want: %d", n, 2)
	}
}
