package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
)

const (
	// Number of cache data store will be 2 ^ concurrencyLevel.
	concurrencyLevel = 2
	segmentCount     = 1 << concurrencyLevel
	segmentMask      = segmentCount - 1
)

// entry stores cached entry key and value.
type entry struct {
	key   Key
	value atomic.Value // Store value

	// hash is the hash value of this entry key
	hash uint64
	// accessTime is the last time this entry was accessed.
	accessTime int64
	// writeTime is the last time this entry was updated.
	writeTime int64

	// FIXME: More efficient way to store boolean flags
	invalidated uint32
	loading     uint32

	// These properties are managed by only cache policy so do not need atomic access.

	// accessList is the list (ordered by access time) this entry is currently in.
	accessList *list.Element
	// listID is ID of the list which this entry is currently in.
	listID uint8
}

func newEntry(k Key, v Value, h uint64) *entry {
	en := &entry{
		key:  k,
		hash: h,
	}
	en.setValue(v)
	return en
}

func (e *entry) getValue() Value {
	return e.value.Load().(Value)
}

func (e *entry) setValue(v Value) {
	e.value.Store(v)
}

func (e *entry) getAccessTime() int64 {
	return atomic.LoadInt64(&e.accessTime)
}

func (e *entry) setAccessTime(v int64) {
	atomic.StoreInt64(&e.accessTime, v)
}

func (e *entry) getWriteTime() int64 {
	return atomic.LoadInt64(&e.writeTime)
}

func (e *entry) setWriteTime(v int64) {
	atomic.StoreInt64(&e.writeTime, v)
}

func (e *entry) getLoading() bool {
	return atomic.LoadUint32(&e.loading) != 0
}

func (e *entry) setLoading(v bool) bool {
	if v {
		return atomic.CompareAndSwapUint32(&e.loading, 0, 1)
	}
	return atomic.CompareAndSwapUint32(&e.loading, 1, 0)
}

func (e *entry) getInvalidated() bool {
	return atomic.LoadUint32(&e.invalidated) != 0
}

func (e *entry) setInvalidated(v bool) {
	if v {
		atomic.StoreUint32(&e.invalidated, 1)
	} else {
		atomic.StoreUint32(&e.invalidated, 0)
	}
}

// getEntry returns the entry attached to the given list element.
func getEntry(el *list.Element) *entry {
	return el.Value.(*entry)
}

// event is the cache event (add, hit or delete).
type event uint8

const (
	eventAdd event = iota
	eventHit
	eventDelete
)

type entryEvent struct {
	entry *entry
	event event
}

// cache is a data structure for cache entries.
type cache struct {
	segs [segmentCount]sync.Map // map[Key]*entry
	size int64
}

func (c *cache) get(k Key, h uint64) *entry {
	seg := c.segment(h)
	v, ok := seg.Load(k)
	if ok {
		return v.(*entry)
	}
	return nil
}

func (c *cache) getOrSet(v *entry) *entry {
	seg := c.segment(v.hash)
	en, ok := seg.LoadOrStore(v.key, v)
	if ok {
		return en.(*entry)
	}
	atomic.AddInt64(&c.size, 1)
	return nil
}

func (c *cache) delete(v *entry) {
	seg := c.segment(v.hash)
	seg.Delete(v.key)
	atomic.AddInt64(&c.size, -1)
}

func (c *cache) len() int {
	return int(atomic.LoadInt64(&c.size))
}

func (c *cache) walk(fn func(*entry)) {
	for i := range c.segs {
		c.segs[i].Range(func(k, v interface{}) bool {
			fn(v.(*entry))
			return true
		})
	}
}

func (c *cache) segment(h uint64) *sync.Map {
	return &c.segs[h&segmentMask]
}

// policy is a cache policy.
type policy interface {
	// init initializes the policy.
	init(cache *cache, maximumSize int)
	// add adds new entry and returns evicted entry if needed.
	add(entry *entry) *entry
	// hit marks then entry recently accessed.
	hit(entry *entry)
	// remove removes the entry.
	remove(entry *entry) *entry
	// walkAccess iterates all entries by their access time.
	walkAccess(func(entry *entry) bool)
}

func newPolicy(name string) policy {
	switch name {
	case "", "slru":
		return &slruCache{}
	case "lru":
		return &lruCache{}
	case "tinylfu":
		return &tinyLFU{}
	default:
		panic("cache: unsupported policy " + name)
	}
}
