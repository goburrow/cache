package cache

import (
	"fmt"
	"testing"
)

type tinyLFUTest struct {
	c   cache
	lfu tinyLFU
	t   *testing.T
}

func (t *tinyLFUTest) assertCap(n int) {
	if t.lfu.lru.cap+t.lfu.slru.protectedCap+t.lfu.slru.probationCap != n {
		t.t.Helper()
		t.t.Fatalf("unexpected lru.cap: %d, slru.cap: %d/%d",
			t.lfu.lru.cap, t.lfu.slru.protectedCap, t.lfu.slru.probationCap)
	}
}

func (t *tinyLFUTest) assertLen(admission, protected, probation int) {
	sz := cacheSize(&t.c)
	az := t.lfu.lru.ls.Len()
	tz := t.lfu.slru.protectedLs.Len()
	bz := t.lfu.slru.probationLs.Len()
	if sz != admission+protected+probation || az != admission || tz != protected || bz != probation {
		t.t.Helper()
		t.t.Fatalf("unexpected data length: cache=%d admission=%d protected=%d probation=%d, want: %d %d %d",
			sz, az, tz, bz, admission, protected, probation)
	}
}

func (t *tinyLFUTest) assertEntry(en *entry, k int, v string, id uint8) {
	ak := en.key.(int)
	av := en.getValue().(string)
	if ak != k || av != v || en.listID != id {
		t.t.Helper()
		t.t.Fatalf("unexpected entry: %+v, want: {key: %d, value: %s, listID: %d}",
			en, k, v, id)
	}
}

func (t *tinyLFUTest) assertLRUEntry(k int, id uint8) {
	en := t.c.get(k, sum(k))
	if en == nil {
		t.t.Helper()
		t.t.Fatalf("entry not found in cache: key=%v", k)
	}
	ak := en.key.(int)
	av := en.getValue().(string)
	v := fmt.Sprintf("%d", k)
	if ak != k || av != v || en.listID != id {
		t.t.Helper()
		t.t.Fatalf("unexpected entry: %+v, want: {key: %d, value: %s, listID: %d}",
			en, k, v, id)
	}
}

func TestTinyLFU(t *testing.T) {
	s := tinyLFUTest{t: t}
	s.lfu.init(&s.c, 200)
	s.assertCap(200)
	s.lfu.slru.protectedCap = 2
	s.lfu.slru.probationCap = 1

	en := make([]*entry, 10)
	for i := range en {
		en[i] = newEntry(i, fmt.Sprintf("%d", i), sum(i))
	}
	for i := 0; i < 5; i++ {
		remEn := s.lfu.write(en[i])
		if remEn != nil {
			t.Fatalf("unexpected entry removed: %+v", remEn)
		}
	}
	// 4 3 | - | 2 1 0
	s.assertLen(2, 0, 3)
	s.assertLRUEntry(4, admissionWindow)
	s.assertLRUEntry(3, admissionWindow)
	s.assertLRUEntry(2, probationSegment)
	s.assertLRUEntry(1, probationSegment)
	s.assertLRUEntry(0, probationSegment)

	s.lfu.access(en[1])
	s.lfu.access(en[2])
	// 4 3 | 2 1 | 0
	s.assertLen(2, 2, 1)
	s.assertLRUEntry(2, protectedSegment)
	s.assertLRUEntry(1, protectedSegment)
	s.assertLRUEntry(0, probationSegment)

	remEn := s.lfu.write(en[5])
	// 5 4 | 2 1 | 0
	if remEn == nil {
		t.Fatalf("expect an entry removed when adding %+v", en[5])
	}
	s.assertEntry(remEn, 3, "3", admissionWindow)

	s.lfu.access(en[4])
	s.lfu.access(en[5])
	remEn = s.lfu.write(en[6])
	// 6 5 | 2 1 | 4
	if remEn == nil {
		t.Fatalf("expect an entry removed when adding %+v", en[6])
	}
	s.assertLen(2, 2, 1)
	s.assertEntry(remEn, 0, "0", probationSegment)
	n := s.lfu.estimate(en[1].hash)
	if n != 2 {
		t.Fatalf("unexpected estimate: %d %+v", n, en[1])
	}
	s.lfu.access(en[2])
	s.lfu.access(en[2])
	n = s.lfu.estimate(en[2].hash)
	if n != 4 {
		t.Fatalf("unexpected estimate: %d %+v", n, en[2])
	}
}
