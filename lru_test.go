package cache

import (
	"container/list"
	"fmt"
	"runtime"
	"testing"
)

func assertLRULen(t *testing.T, c *cache, n int) {
	if len(c.data) != n {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d\nunexpected data length: %d\nwant: %d",
			file, line, len(c.data), n)
	}
}

func assertLRUEntry(t *testing.T, en *entry, k int, v string, id listID) {
	if en.key.(int) != k || en.value.(string) != v || en.listID != id {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d\nunexpected entry: %+v\nwant: {key: %d, value: %s, listID: %d}",
			file, line, en, k, v, id)
	}
}

func TestLRU(t *testing.T) {
	c := &cache{
		data: make(map[Key]*list.Element),
	}
	l := &lruCache{}
	l.init(c, 3)

	en := make([]*entry, 4)
	for i := 0; i < len(en); i++ {
		en[i] = &entry{key: i, value: fmt.Sprintf("%d", i)}
	}
	// 0
	remEn := l.add(en[0])
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %v", remEn)
	}
	assertLRULen(t, c, 1)
	el0 := c.data[en[0].key]
	en0 := getEntry(el0)
	assertLRUEntry(t, en0, 0, "0", 0)
	// 1 0
	remEn = l.add(en[1])
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %v", remEn)
	}
	assertLRULen(t, c, 2)
	el1 := c.data[en[1].key]
	en1 := getEntry(el1)
	assertLRUEntry(t, en1, 1, "1", 0)
	// 0 1
	l.hit(el0)
	// 2 0 1
	remEn = l.add(en[2])
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %+v", remEn)
	}
	assertLRULen(t, c, 3)
	// 3 2 0
	remEn = l.add(en[3])
	assertLRUEntry(t, remEn, 1, "1", 0)
	assertLRULen(t, c, 3)
	el3 := c.data[en[3].key]
	en3 := getEntry(el3)
	assertLRUEntry(t, en3, 3, "3", 0)
	// 3 4
	el2 := c.data[en[2].key]
	remEn = l.remove(el2)
	assertLRUEntry(t, remEn, 2, "2", 0)
	assertLRULen(t, c, 2)
}

func TestSegmentedLRU(t *testing.T) {
	c := &cache{
		data: make(map[Key]*list.Element),
	}
	l := &slruCache{}
	l.init(c, 3)
	l.probationCap = 1
	l.protectedCap = 2

	en := make([]*entry, 5)
	for i := 0; i < len(en); i++ {
		en[i] = &entry{key: i, value: fmt.Sprintf("%d", i)}
	}
	// - | 0
	remEn := l.add(en[0])
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %v", remEn)
	}
	assertLRULen(t, c, 1)
	el0 := c.data[en[0].key]
	en0 := getEntry(el0)
	assertLRUEntry(t, en0, 0, "0", probationSegment)
	// - | 1 0
	remEn = l.add(en[1])
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %v", remEn)
	}
	assertLRULen(t, c, 2)
	el1 := c.data[en[1].key]
	en1 := getEntry(el1)
	assertLRUEntry(t, en1, 1, "1", probationSegment)
	// 1 | 0
	l.hit(el1)
	el1 = c.data[en[1].key]
	en1 = getEntry(el1)
	assertLRUEntry(t, en1, 1, "1", protectedSegment)
	assertLRUEntry(t, en0, 0, "0", probationSegment)
	// 0 1 | -
	l.hit(el0)
	el0 = c.data[en[0].key]
	en0 = getEntry(el0)
	assertLRUEntry(t, en0, 0, "0", protectedSegment)
	// 0 1 | 2
	remEn = l.add(en[2])
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %+v", remEn)
	}
	assertLRULen(t, c, 3)
	// 0 1 | 3
	remEn = l.add(en[3])
	assertLRUEntry(t, remEn, 2, "2", probationSegment)
	assertLRULen(t, c, 3)
	el3 := c.data[en[3].key]
	en3 := getEntry(el3)
	assertLRUEntry(t, en3, 3, "3", probationSegment)
	// 3 0 | 1
	l.hit(el3)
	el3 = c.data[en[3].key]
	en3 = getEntry(el3)
	assertLRUEntry(t, en3, 3, "3", protectedSegment)
	// 3 0 | 4
	remEn = l.add(en[4])
	assertLRUEntry(t, remEn, 1, "1", probationSegment)
	// 3 | 4
	el0 = c.data[en[0].key]
	remEn = l.remove(el0)
	assertLRUEntry(t, remEn, 0, "0", protectedSegment)
	assertLRULen(t, c, 2)
}
