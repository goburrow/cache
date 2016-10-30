package cache

import (
	"container/list"
	"fmt"
	"runtime"
	"testing"
)

func assertLRULen(t *testing.T, c *cache, n int) {
	_, file, line, _ := runtime.Caller(1)
	if len(c.data) != n {
		t.Fatalf("%s:%d\nunexpected data length: %d\nwant: %d",
			file, line, len(c.data), n)
	}
}

func assertLRUEntry(t *testing.T, en *entry, k int, v string) {
	_, file, line, _ := runtime.Caller(1)
	if en.key.(int) != k || en.value.(string) != v {
		t.Fatalf("%s:%d\nunexpected entry: %+v\nwant: {key: %d, value: %s}",
			file, line, en, k, v)
	}
}

func TestLRU(t *testing.T) {
	c := &cache{
		data: make(map[Key]*list.Element),
		cap:  3,
	}
	l := &lruCache{}
	l.init(c)

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
	assertLRUEntry(t, en0, 0, "0")
	// 1 0
	remEn = l.add(en[1])
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %v", remEn)
	}
	assertLRULen(t, c, 2)
	el1 := c.data[en[1].key]
	en1 := getEntry(el1)
	assertLRUEntry(t, en1, 1, "1")
	// 0 1
	l.access(el0)
	// 2 0 1
	remEn = l.add(en[2])
	if remEn != nil {
		t.Fatalf("unexpected entry removed: %+v", remEn)
	}
	assertLRULen(t, c, 3)
	// 3 2 0
	remEn = l.add(en[3])
	assertLRUEntry(t, remEn, 1, "1")
	assertLRULen(t, c, 3)
	el3 := c.data[en[3].key]
	en3 := getEntry(el3)
	assertLRUEntry(t, en3, 3, "3")
	// 3 4
	el2 := c.data[en[2].key]
	remEn = l.remove(el2)
	assertLRUEntry(t, remEn, 2, "2")
	assertLRULen(t, c, 2)
}
