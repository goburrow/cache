package cache

import (
	"container/list"
	"fmt"
	"testing"
)

func TestTinyLFU(t *testing.T) {
	c := &cache{
		data: make(map[Key]*list.Element),
	}
	l := tinyLFU{
		doorkeeperEnabled: true,
	}
	l.init(c, 200)
	if l.lru.cap+l.slru.protectedCap+l.slru.probationCap != 200 {
		t.Fatalf("unexpected lru.cap: %d, slru.cap: %d/%d",
			l.lru.cap, l.slru.protectedCap, l.slru.probationCap)
	}
	l.slru.protectedCap = 2
	l.slru.probationCap = 1

	en := make([]*entry, 10)
	for i := 0; i < len(en); i++ {
		en[i] = &entry{
			key:   i,
			value: fmt.Sprintf("%d", i),
			hash:  sum(i),
		}
	}
	// 4 3 | - | 2 1 0
	for i := 0; i < 5; i++ {
		remEn := l.add(en[i])
		if remEn != nil {
			t.Fatalf("unexpected entry removed: %+v", remEn)
		}
	}
	en4 := getEntry(c.data[en[4].key])
	assertLRUEntry(t, en4, 4, "4", admissionWindow)

	en3 := getEntry(c.data[en[3].key])
	assertLRUEntry(t, en3, 3, "3", admissionWindow)

	en0 := getEntry(c.data[en[0].key])
	assertLRUEntry(t, en0, 0, "0", probationSegment)

	// 4 3 | 2 1 | 0
	l.hit(c.data[en[1].key])
	l.hit(c.data[en[2].key])
	en1 := getEntry(c.data[en[1].key])
	assertLRUEntry(t, en1, 1, "1", protectedSegment)
	en2 := getEntry(c.data[en[2].key])
	assertLRUEntry(t, en2, 2, "2", protectedSegment)

	// 5 4 | 2 1 | 0
	remEn := l.add(en[5])
	if remEn == nil {
		t.Fatalf("expect an entry removed when adding %+v", en[5])
	}
	assertLRUEntry(t, remEn, 3, "3", admissionWindow)

	l.hit(c.data[en[4].key])
	l.hit(c.data[en[5].key])
	// 6 5 | 2 1 | 4
	remEn = l.add(en[6])
	if remEn == nil {
		t.Fatalf("expect an entry removed when adding %+v", en[6])
	}
	assertLRUEntry(t, remEn, 0, "0", probationSegment)
	n := l.estimate(en[1].hash)
	if n != 2 {
		t.Fatalf("unexpected estimate: %d %+v", n, en[1])
	}
	l.hit(c.data[en[2].key])
	l.hit(c.data[en[2].key])
	n = l.estimate(en[2].hash)
	if n != 4 {
		t.Fatalf("unexpected estimate: %d %+v", n, en[2])
	}
}
