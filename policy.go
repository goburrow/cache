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
type entry[Key comparable, Value any] struct {
	// Structs with first field align to 64 bits will also be aligned to 64.
	// https://golang.org/pkg/sync/atomic/#pkg-note-BUG

	// hash is the hash value of this entry key
	hash uint64
	// accessTime is the last time this entry was accessed.
	accessTime int64 // Access atomically - must be aligned on 32-bit
	// writeTime is the last time this entry was updated.
	writeTime int64 // Access atomically - must be aligned on 32-bit

	// FIXME: More efficient way to store boolean flags
	invalidated int32
	loading     int32

	key   Key
	value atomic.Value // Store value

	// These properties are managed by only cache policy so do not need atomic access.

	// accessList is the list (ordered by access time) this entry is currently in.
	accessList *list.Element
	// writeList is the list (ordered by write time) this entry is currently in.
	writeList *list.Element
	// listID is ID of the list which this entry is currently in.
	listID uint8
}

func newEntry[Key comparable, Value any](k Key, v Value, h uint64) *entry[Key, Value] {
	en := &entry[Key, Value]{
		key:  k,
		hash: h,
	}
	en.setValue(v)
	return en
}

func (e *entry[Key, Value]) getValue() Value {
	return e.value.Load().(Value)
}

func (e *entry[Key, Value]) setValue(v Value) {
	e.value.Store(v)
}

func (e *entry[Key, Value]) getAccessTime() int64 {
	return atomic.LoadInt64(&e.accessTime)
}

func (e *entry[Key, Value]) setAccessTime(v int64) {
	atomic.StoreInt64(&e.accessTime, v)
}

func (e *entry[Key, Value]) getWriteTime() int64 {
	return atomic.LoadInt64(&e.writeTime)
}

func (e *entry[Key, Value]) setWriteTime(v int64) {
	atomic.StoreInt64(&e.writeTime, v)
}

func (e *entry[Key, Value]) getLoading() bool {
	return atomic.LoadInt32(&e.loading) != 0
}

func (e *entry[Key, Value]) setLoading(v bool) bool {
	if v {
		return atomic.CompareAndSwapInt32(&e.loading, 0, 1)
	}
	return atomic.CompareAndSwapInt32(&e.loading, 1, 0)
}

func (e *entry[Key, Value]) getInvalidated() bool {
	return atomic.LoadInt32(&e.invalidated) != 0
}

func (e *entry[Key, Value]) setInvalidated(v bool) {
	if v {
		atomic.StoreInt32(&e.invalidated, 1)
	} else {
		atomic.StoreInt32(&e.invalidated, 0)
	}
}

// getEntry returns the entry attached to the given list element.
func getEntry[Key comparable, Value any](el *list.Element) *entry[Key, Value] {
	return el.Value.(*entry[Key, Value])
}

// event is the cache event (add, hit or delete).
type event uint8

const (
	eventWrite event = iota
	eventAccess
	eventDelete
	eventClose
)

type entryEvent[Key comparable, Value any] struct {
	entry *entry[Key, Value]
	event event
}

// cache is a data structure for cache entries.
type cache[Key comparable, Value any] struct {
	size int64                  // Access atomically - must be aligned on 32-bit
	segs [segmentCount]sync.Map // map[Key]*entry
}

func (c *cache[Key, Value]) get(k Key, h uint64) *entry[Key, Value] {
	seg := c.segment(h)
	v, ok := seg.Load(k)
	if ok {
		return v.(*entry[Key, Value])
	}
	return nil
}

func (c *cache[Key, Value]) getOrSet(v *entry[Key, Value]) *entry[Key, Value] {
	seg := c.segment(v.hash)
	en, ok := seg.LoadOrStore(v.key, v)
	if ok {
		return en.(*entry[Key, Value])
	}
	atomic.AddInt64(&c.size, 1)
	return nil
}

func (c *cache[Key, Value]) delete(v *entry[Key, Value]) {
	seg := c.segment(v.hash)
	seg.Delete(v.key)
	atomic.AddInt64(&c.size, -1)
}

func (c *cache[Key, Value]) len() int {
	return int(atomic.LoadInt64(&c.size))
}

func (c *cache[Key, Value]) walk(fn func(*entry[Key, Value])) {
	for i := range c.segs {
		c.segs[i].Range(func(k, v any) bool {
			fn(v.(*entry[Key, Value]))
			return true
		})
	}
}

func (c *cache[Key, Value]) segment(h uint64) *sync.Map {
	return &c.segs[h&segmentMask]
}

// policy is a cache policy.
type policy[Key comparable, Value any] interface {
	// init initializes the policy.
	init(cache *cache[Key, Value], maximumSize int)
	// write handles Write event for the entry.
	// It adds new entry and returns evicted entry if needed.
	write(entry *entry[Key, Value]) *entry[Key, Value]
	// access handles Access event for the entry.
	// It marks then entry recently accessed.
	access(entry *entry[Key, Value])
	// remove removes the entry.
	remove(entry *entry[Key, Value]) *entry[Key, Value]
	// iterate iterates all entries by their access time.
	iterate(func(entry *entry[Key, Value]) bool)
}

func newPolicy[Key comparable, Value any](name string) policy[Key, Value] {
	switch name {
	case "", "slru":
		return &slruCache[Key, Value]{}
	case "lru":
		return &lruCache[Key, Value]{}
	case "tinylfu":
		return &tinyLFU[Key, Value]{}
	default:
		panic("cache: unsupported policy " + name)
	}
}

// recencyQueue manages cache entries by write time.
type recencyQueue[Key comparable, Value any] struct {
	ls list.List
}

func (w *recencyQueue[Key, Value]) init(cache *cache[Key, Value], maximumSize int) {
	w.ls.Init()
}

func (w *recencyQueue[Key, Value]) write(en *entry[Key, Value]) *entry[Key, Value] {
	if en.writeList == nil {
		en.writeList = w.ls.PushFront(en)
	} else {
		w.ls.MoveToFront(en.writeList)
	}
	return nil
}

func (w *recencyQueue[Key, Value]) access(en *entry[Key, Value]) {
}

func (w *recencyQueue[Key, Value]) remove(en *entry[Key, Value]) *entry[Key, Value] {
	if en.writeList == nil {
		return en
	}
	w.ls.Remove(en.writeList)
	en.writeList = nil
	return en
}

func (w *recencyQueue[Key, Value]) iterate(fn func(en *entry[Key, Value]) bool) {
	iterateListFromBack(&w.ls, fn)
}

type discardingQueue[Key comparable, Value any] struct{}

func (discardingQueue[Key, Value]) init(cache *cache[Key, Value], maximumSize int) {
}

func (discardingQueue[Key, Value]) write(en *entry[Key, Value]) *entry[Key, Value] {
	return nil
}

func (discardingQueue[Key, Value]) access(en *entry[Key, Value]) {
}

func (discardingQueue[Key, Value]) remove(en *entry[Key, Value]) *entry[Key, Value] {
	return en
}

func (discardingQueue[Key, Value]) iterate(fn func(en *entry[Key, Value]) bool) {
}

func iterateListFromBack[Key comparable, Value any](ls *list.List, fn func(en *entry[Key, Value]) bool) {
	for el := ls.Back(); el != nil; {
		en := getEntry[Key, Value](el)
		prev := el.Prev() // Get Prev as fn can delete the entry.
		if !fn(en) {
			return
		}
		el = prev
	}
}
