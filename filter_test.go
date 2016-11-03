package cache

import "testing"

func TestBloomFilter(t *testing.T) {
	const numIns = 100000
	f := bloomFilter{}
	f.init(numIns, 0.01)

	var i uint64
	for i = 0; i < numIns; i += 2 {
		existed := f.put(i)
		if existed {
			t.Fatalf("unexpected put(%d): %v, want: false", i, existed)
		}
	}
	for i = 0; i < numIns; i += 2 {
		existed := f.contains(i)
		if !existed {
			t.Fatalf("unexpected contains(%d): %v, want: true", i, existed)
		}
	}
	for i = 1; i < numIns; i += 2 {
		existed := f.contains(i)
		if existed {
			t.Fatalf("unexpected contains(%d): %v, want: false", i, existed)
		}
	}
	for i = 0; i < numIns; i += 2 {
		existed := f.put(i)
		if !existed {
			t.Fatalf("unexpected put(%d): %v, want: true", i, existed)
		}
	}
}
