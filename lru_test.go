package cache

import (
	"fmt"
	"testing"
)

type lruTest struct {
	c    cache
	lru  lruCache
	slru slruCache
	t    *testing.T
}

func (t *lruTest) assertLRULen(n int) {
	sz := cacheSize(&t.c)
	lz := t.lru.ls.Len()
	if sz != n || lz != n {
		t.t.Helper()
		t.t.Fatalf("unexpected data length: cache=%d list=%d, want: %d", sz, lz, n)
	}
}

func (t *lruTest) assertSLRULen(protected, probation int) {
	sz := cacheSize(&t.c)
	tz := t.slru.protectedLs.Len()
	bz := t.slru.probationLs.Len()
	if sz != protected+probation || tz != protected || bz != probation {
		t.t.Helper()
		t.t.Fatalf("unexpected data length: cache=%d protected=%d probation=%d, want: %d %d", sz, tz, bz, protected, probation)
	}
}

func (t *lruTest) assertEntry(en *entry, k int, v string, id uint8) {
	if en == nil {
		t.t.Helper()
		t.t.Fatalf("unexpected entry: %v", en)
	}
	ak := en.key.(int)
	av := en.getValue().(string)
	if ak != k || av != v || en.listID != id {
		t.t.Helper()
		t.t.Fatalf("unexpected entry: %+v, want: {key: %d, value: %s, listID: %d}",
			en, k, v, id)
	}
}

func (t *lruTest) assertLRUEntry(k int) {
	en := t.c.get(k, 0)
	if en == nil {
		t.t.Helper()
		t.t.Fatalf("entry not found in cache: key=%v", k)
	}
	ak := en.key.(int)
	av := en.getValue().(string)
	v := fmt.Sprintf("%d", k)
	if ak != k || av != v || en.listID != 0 {
		t.t.Helper()
		t.t.Fatalf("unexpected entry: %+v, want: {key: %v, value: %v, listID: %v}", en, k, v, 0)
	}
}

func (t *lruTest) assertSLRUEntry(k int, id uint8) {
	en := t.c.get(k, 0)
	if en == nil {
		t.t.Helper()
		t.t.Fatalf("entry not found in cache: key=%v", k)
	}
	ak := en.key.(int)
	av := en.getValue().(string)
	v := fmt.Sprintf("%d", k)
	if ak != k || av != v || en.listID != id {
		t.t.Helper()
		t.t.Fatalf("unexpected entry: %+v, want: {key: %v, value: %v, listID: %v}", en, k, v, id)
	}
}

func TestLRU(t *testing.T) {
	s := lruTest{t: t}
	s.lru.init(&s.c, 3)

	en := createLRUEntries(4)
	remEn := s.lru.write(en[0])
	// 0
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %v", remEn)
	}
	s.assertLRULen(1)
	s.assertLRUEntry(0)
	remEn = s.lru.write(en[1])
	// 1 0
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %v", remEn)
	}
	s.assertLRULen(2)
	s.assertLRUEntry(1)
	s.assertLRUEntry(0)

	s.lru.access(en[0])
	// 0 1

	remEn = s.lru.write(en[2])
	// 2 0 1
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %+v", remEn)
	}
	s.assertLRULen(3)

	remEn = s.lru.write(en[3])
	// 3 2 0
	s.assertEntry(remEn, 1, "1", 0)
	s.assertLRULen(3)
	s.assertLRUEntry(3)
	s.assertLRUEntry(2)
	s.assertLRUEntry(0)

	remEn = s.lru.remove(en[2])
	// 3 0
	s.assertEntry(remEn, 2, "2", 0)
	s.assertLRULen(2)
	s.assertLRUEntry(3)
	s.assertLRUEntry(0)
}

func TestLRUWalk(t *testing.T) {
	s := lruTest{t: t}
	s.lru.init(&s.c, 5)

	entries := createLRUEntries(6)
	for _, e := range entries {
		s.lru.write(e)
	}
	// 5 4 3 2 1
	found := ""
	s.lru.iterate(func(en *entry) bool {
		found += en.getValue().(string) + " "
		return true
	})
	if found != "1 2 3 4 5 " {
		t.Fatalf("unexpected entries: %v", found)
	}
	s.lru.access(entries[1])
	s.lru.access(entries[5])
	s.lru.access(entries[3])
	// 3 5 1 4 2
	found = ""
	s.lru.iterate(func(en *entry) bool {
		found += en.getValue().(string) + " "
		if en.key.(int)%2 == 0 {
			s.lru.remove(en)
		}
		return en.key.(int) != 5
	})
	if found != "2 4 1 5 " {
		t.Fatalf("unexpected entries: %v", found)
	}
	s.assertLRULen(3)
	s.assertLRUEntry(3)
	s.assertLRUEntry(5)
	s.assertLRUEntry(1)
}

func TestSegmentedLRU(t *testing.T) {
	s := lruTest{t: t}
	s.slru.init(&s.c, 3)
	s.slru.probationCap = 1
	s.slru.protectedCap = 2

	en := createLRUEntries(5)

	remEn := s.slru.write(en[0])
	// - | 0
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %v", remEn)
	}
	s.assertSLRULen(0, 1)
	s.assertSLRUEntry(0, probationSegment)

	remEn = s.slru.write(en[1])
	// - | 1 0
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %v", remEn)
	}
	s.assertSLRULen(0, 2)
	s.assertSLRUEntry(1, probationSegment)

	s.slru.access(en[1])
	// 1 | 0
	s.assertSLRULen(1, 1)
	s.assertSLRUEntry(1, protectedSegment)
	s.assertSLRUEntry(0, probationSegment)

	s.slru.access(en[0])
	// 0 1 | -
	s.assertSLRULen(2, 0)
	s.assertSLRUEntry(0, protectedSegment)
	s.assertSLRUEntry(1, protectedSegment)

	remEn = s.slru.write(en[2])
	// 0 1 | 2
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %+v", remEn)
	}
	s.assertSLRULen(2, 1)
	s.assertSLRUEntry(2, probationSegment)

	remEn = s.slru.write(en[3])
	// 0 1 | 3
	s.assertSLRULen(2, 1)
	s.assertEntry(remEn, 2, "2", probationSegment)
	s.assertSLRUEntry(3, probationSegment)

	s.slru.access(en[3])
	// 3 0 | 1
	s.assertSLRULen(2, 1)
	s.assertSLRUEntry(3, protectedSegment)

	remEn = s.slru.write(en[4])
	// 3 0 | 4
	s.assertSLRULen(2, 1)
	s.assertEntry(remEn, 1, "1", probationSegment)

	remEn = s.slru.remove(en[0])
	// 3 | 4
	s.assertEntry(remEn, 0, "0", protectedSegment)
	s.assertSLRULen(1, 1)
	s.assertSLRUEntry(3, protectedSegment)
	s.assertSLRUEntry(4, probationSegment)
}

func TestSLRUWalk(t *testing.T) {
	s := lruTest{t: t}
	s.slru.init(&s.c, 6)

	entries := createLRUEntries(10)
	for _, e := range entries {
		s.slru.write(e)
	}
	// | 9 8 7 6 5 4
	found := ""
	s.slru.iterate(func(en *entry) bool {
		found += en.getValue().(string) + " "
		return true
	})
	if found != "4 5 6 7 8 9 " {
		t.Fatalf("unexpected entries: %v", found)
	}
	s.slru.access(entries[7])
	s.slru.access(entries[5])
	s.slru.access(entries[8])
	// 8 5 7 | 9 6 4
	found = ""
	s.slru.iterate(func(en *entry) bool {
		found += en.getValue().(string) + " "
		if en.key.(int)%2 == 0 {
			s.slru.remove(en)
		}
		return en.key.(int) != 6
	})
	if found != "7 5 8 4 6 " {
		t.Fatalf("unexpected entries: %v", found)
	}
	s.assertSLRULen(2, 1)
	s.assertSLRUEntry(5, protectedSegment)
	s.assertSLRUEntry(7, protectedSegment)
	s.assertSLRUEntry(9, probationSegment)
}

func createLRUEntries(n int) []*entry {
	en := make([]*entry, n)
	for i := range en {
		en[i] = newEntry(i, fmt.Sprintf("%d", i), 0 /* unused */)
	}
	return en
}
