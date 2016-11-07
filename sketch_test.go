package cache

import "testing"

func TestCountMinSketch(t *testing.T) {
	const max = 15
	cm := &countMinSketch{}
	cm.init(max)
	for i := 0; i < max; i++ {
		// Increase value at i j times
		for j := i; j > 0; j-- {
			cm.add(uint64(i))
		}
	}
	for i := 0; i < max; i++ {
		n := cm.estimate(uint64(i))
		if int(n) != i {
			t.Fatalf("unexpected estimate(%d): %d, want: %d", i, n, i)
		}
	}
	cm.reset()
	for i := 0; i < max; i++ {
		n := cm.estimate(uint64(i))
		if int(n) != i/2 {
			t.Fatalf("unexpected estimate(%d): %d, want: %d", i, n, i/2)
		}
	}
	cm.reset()
	for i := 0; i < max; i++ {
		n := cm.estimate(uint64(i))
		if int(n) != i/4 {
			t.Fatalf("unexpected estimate(%d): %d, want: %d", i, n, i/4)
		}
	}
	for i := 0; i < 100; i++ {
		cm.add(1)
	}
	n := cm.estimate(1)
	if n != 15 {
		t.Fatalf("unexpected estimate(%d): %d, want: %d", 1, n, 15)
	}
}

func BenchmarkCountMinSketchReset(b *testing.B) {
	cm := &countMinSketch{}
	cm.init(1<<15 - 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cm.add(0xCAFECAFECAFECAFE)
		cm.reset()
	}
}
