package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// Default maximum number of cache entries.
	defaultMaxSize = 1<<31 - 1
	// Buffer size of entry channels
	chanBufSize = 1
	// Maximum number of entries to be drained in a single clean up.
	drainMax       = 16
	drainThreshold = 64
)

// currentTime is an alias for time.Now, used for testing.
var currentTime = time.Now

// entry stores cached entry key and value.
type entry struct {
	key   Key
	value Value

	// accessed is the last time this entry was accessed.
	accessed time.Time
	// updated is the last time this entry was updated.
	updated time.Time
}

func getEntry(el *list.Element) *entry {
	return el.Value.(*entry)
}

func setEntry(el *list.Element, en *entry) {
	el.Value = en
}

// localCache is an asynchronous LRU cache.
type localCache struct {
	maximumSize int

	expireAfterAccess time.Duration
	refreshAfterWrite time.Duration

	onInsertion Func
	onRemoval   Func

	loader LoaderFunc
	stats  Stats

	cacheMu sync.RWMutex
	cache   map[Key]*list.Element

	entries     list.List
	addEntry    chan *entry
	accessEntry chan *list.Element
	deleteEntry chan *list.Element

	// readCount is a counter of the number of reads since the last write.
	readCount int32

	closeCh chan struct{}
}

// newLocalCache returns a default localCache
func newLocalCache() *localCache {
	c := &localCache{
		maximumSize: defaultMaxSize,
		cache:       make(map[Key]*list.Element),

		addEntry:    make(chan *entry, chanBufSize),
		accessEntry: make(chan *list.Element, chanBufSize),
		deleteEntry: make(chan *list.Element, chanBufSize),
	}
	c.entries.Init()
	return c
}

func (c *localCache) start() {
	c.closeCh = make(chan struct{})
	go c.processEntries()
}

// Close is for implementing io.Closer.
// It always return nil error.
func (c *localCache) Close() error {
	close(c.closeCh)
	return nil
}

// GetIfPresent gets cached value from entries list and updates
// last access time for the entry if it is found.
func (c *localCache) GetIfPresent(k Key) (Value, bool) {
	c.cacheMu.RLock()
	el, hit := c.cache[k]
	c.cacheMu.RUnlock()
	if !hit {
		c.accessEntry <- nil
		c.stats.AddMissCount(1)
		return nil, false
	}
	v := getEntry(el).value
	c.accessEntry <- el
	c.stats.AddHitCount(1)
	return v, true
}

// Put adds new entry to entries list.
func (c *localCache) Put(k Key, v Value) {
	c.cacheMu.RLock()
	el, hit := c.cache[k]
	c.cacheMu.RUnlock()
	if hit {
		// Update list element value
		getEntry(el).value = v
		c.accessEntry <- el
	} else {
		en := &entry{
			key:   k,
			value: v,
		}
		c.addEntry <- en
	}
}

// Invalidate removes the entry associated with key k.
func (c *localCache) Invalidate(k Key) {
	c.cacheMu.RLock()
	el, hit := c.cache[k]
	c.cacheMu.RUnlock()
	if hit {
		c.deleteEntry <- el
	}
}

// InvalidateAll resets entries list.
func (c *localCache) InvalidateAll() {
	c.deleteEntry <- nil
}

// Get returns value associated with k or call underlying loader to retrieve value
// if it is not in the cache. The returned value is only cached when loader returns
// nil error.
func (c *localCache) Get(k Key) (Value, error) {
	c.cacheMu.RLock()
	el, hit := c.cache[k]
	c.cacheMu.RUnlock()
	if !hit {
		c.stats.AddMissCount(1)
		return c.load(k)
	}
	c.stats.AddHitCount(1)
	en := getEntry(el)
	// Check if this entry needs to be refreshed
	if c.refreshAfterWrite > 0 && en.updated.Before(currentTime().Add(-c.refreshAfterWrite)) {
		return c.refresh(en), nil
	}
	v := en.value
	c.accessEntry <- el
	return v, nil
}

// Stats copies cache stats to t.
func (c *localCache) Stats(t *Stats) {
	c.stats.Copy(t)
}

func (c *localCache) processEntries() {
	for {
		select {
		case <-c.closeCh:
			c.removeAll()
			return
		case en := <-c.addEntry:
			c.add(en)
			c.postWriteCleanup()
		case el := <-c.accessEntry:
			if el != nil {
				c.touch(el)
			}
			c.postReadCleanup()
		case el := <-c.deleteEntry:
			if el == nil {
				c.removeAll()
			} else {
				c.remove(el)
			}
			c.postReadCleanup()
		}
	}
}

func (c *localCache) add(en *entry) {
	en.accessed = currentTime()
	en.updated = en.accessed
	c.cacheMu.Lock()
	el, ok := c.cache[en.key]
	if ok {
		c.cacheMu.Unlock()
		setEntry(el, en)
		c.entries.MoveToFront(el)
	} else {
		var remEn *entry
		if c.maximumSize > 0 && c.entries.Len() >= c.maximumSize {
			// Swap with the oldest one
			el = c.entries.Back()
			if el == nil {
				// This can not happen
				el = c.entries.PushFront(en)
			} else {
				remEn = getEntry(el)
				delete(c.cache, remEn.key)
				setEntry(el, en)
				c.entries.MoveToFront(el)
			}
		} else {
			el = c.entries.PushFront(en)
		}
		c.cache[en.key] = el
		c.cacheMu.Unlock()
		if c.onInsertion != nil {
			c.onInsertion(en.key, en.value)
		}
		if remEn != nil {
			// An entry has been evicted
			c.stats.AddEvictionCount(1)
			if c.onRemoval != nil {
				c.onRemoval(remEn.key, remEn.value)
			}
		}
	}
}

// removeAll remove all entries in the cache.
func (c *localCache) removeAll() {
	c.cacheMu.Lock()
	oldCache := c.cache
	c.cache = make(map[Key]*list.Element)
	c.cacheMu.Unlock()
	c.entries.Init()

	if c.onRemoval != nil {
		for _, el := range oldCache {
			en := getEntry(el)
			c.onRemoval(en.key, en.value)
		}
	}
}

// remove removes the given element from the cache and entries list.
// It also calls onRemoval callback if it is set.
func (c *localCache) remove(el *list.Element) {
	en := getEntry(el)
	c.cacheMu.Lock()
	delete(c.cache, en.key)
	c.cacheMu.Unlock()
	c.entries.Remove(el)

	if c.onRemoval != nil {
		c.onRemoval(en.key, en.value)
	}
}

// touch moves the given element to the top of the entries list.
func (c *localCache) touch(el *list.Element) {
	getEntry(el).accessed = currentTime()
	c.entries.MoveToFront(el)
}

// load uses current loader to retrieve value for k and adds new
// entry to the cache only if loader returns a nil error.
func (c *localCache) load(k Key) (Value, error) {
	if c.loader == nil {
		panic("loader must be set")
	}
	v, err := c.loader(k)
	if err != nil {
		c.stats.AddLoadErrorCount(1)
		return nil, err
	}
	en := &entry{
		key:   k,
		value: v,
	}
	c.addEntry <- en
	c.stats.AddLoadSuccessCount(1)
	return v, nil
}

// refresh reloads value for the given key. If loader returns an error,
// that error will be omitted and current value will be returned.
// Otherwise, the function will returns new value and updates the current
// cache entry.
func (c *localCache) refresh(en *entry) Value {
	if c.loader == nil {
		panic("loader must be set")
	}
	newV, err := c.loader(en.key)
	if err != nil {
		c.stats.AddLoadErrorCount(1)
		return en.value
	}
	en.value = newV
	c.addEntry <- en
	c.stats.AddLoadSuccessCount(1)
	return newV
}

// postReadCleanup is run after entry access/delete event.
func (c *localCache) postReadCleanup() {
	if atomic.AddInt32(&c.readCount, 1) > drainThreshold {
		atomic.StoreInt32(&c.readCount, 0)
		c.expireEntries()
	}
}

// postWriteCleanup is run after entry add event.
func (c *localCache) postWriteCleanup() {
	atomic.StoreInt32(&c.readCount, 0)
	c.expireEntries()
}

// expireEntries removes expired entries.
func (c *localCache) expireEntries() {
	if c.expireAfterAccess <= 0 {
		return
	}
	expire := currentTime().Add(-c.expireAfterAccess)
	for i := drainMax; i > 0; i-- {
		el := c.entries.Back()
		if el == nil {
			// List is empty
			break
		}
		en := getEntry(el)
		if !en.accessed.Before(expire) {
			// Can break since the entries list is sorted by access time
			break
		}
		c.remove(el)
		c.stats.AddEvictionCount(1)
	}
}

// New returns a local in-memory Cache.
func New(options ...Option) Cache {
	c := newLocalCache()
	for _, opt := range options {
		opt(c)
	}
	c.start()
	return c
}

// NewLoadingCache returns a new LoadingCache with given loader function
// and cache options.
func NewLoadingCache(loader LoaderFunc, options ...Option) LoadingCache {
	c := newLocalCache()
	c.loader = loader
	for _, opt := range options {
		opt(c)
	}
	c.start()
	return c
}

// Option add options for default Cache.
type Option func(c *localCache)

// WithMaximumSize returns an Option which sets maximum size for default Cache.
// Any non-positive numbers is considered as unlimited.
func WithMaximumSize(size int) Option {
	return func(c *localCache) {
		c.maximumSize = size
	}
}

// WithRemovalListener returns an Option to set cache to call onRemoval for each
// entry evicted from the cache.
func WithRemovalListener(onRemoval Func) Option {
	return func(c *localCache) {
		c.onRemoval = onRemoval
	}
}

// WithExpireAfterAccess returns an option to expire a cache entry after the
// given duration without being accessed.
func WithExpireAfterAccess(d time.Duration) Option {
	return func(c *localCache) {
		c.expireAfterAccess = d
	}
}

// WithRefreshAfterWrite returns an option to refresh a cache entry after the
// given duration. This option is only applicable for LoadingCache.
func WithRefreshAfterWrite(d time.Duration) Option {
	return func(c *localCache) {
		c.refreshAfterWrite = d
	}
}

// withInsertionListener is used for testing.
func withInsertionListener(onInsertion Func) Option {
	return func(c *localCache) {
		c.onInsertion = onInsertion
	}
}
