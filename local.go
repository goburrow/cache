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
	// Default channel buffer size.
	defaultChanSize = 1

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

	// expire is the expiration time of this entry.
	expire time.Time
}

func getEntry(el *list.Element) *entry {
	return el.Value.(*entry)
}

// localCache is an asynchronous LRU cache.
type localCache struct {
	maximumSize int

	expireAfterAccess time.Duration

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

		addEntry:    make(chan *entry, defaultChanSize),
		accessEntry: make(chan *list.Element, defaultChanSize),
		deleteEntry: make(chan *list.Element, defaultChanSize),
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
	if hit {
		c.stats.AddHitCount(1)
		v := getEntry(el).value
		c.accessEntry <- el
		return v, true
	}
	c.stats.AddMissCount(1)
	c.accessEntry <- nil
	return nil, false
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
	if hit {
		c.stats.AddHitCount(1)
		v := getEntry(el).value
		c.accessEntry <- el
		return v, nil
	}
	c.stats.AddMissCount(1)
	if c.loader == nil {
		panic("loader must be set")
	}
	v, err := c.loader(k)
	if err != nil {
		c.stats.AddLoadErrorCount(1)
		return nil, err
	}
	c.stats.AddLoadSuccessCount(1)
	en := &entry{
		key:   k,
		value: v,
	}
	c.addEntry <- en
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
	if c.expireAfterAccess > 0 {
		en.expire = currentTime().Add(c.expireAfterAccess)
	}
	c.cacheMu.Lock()
	el, ok := c.cache[en.key]
	if ok {
		c.cacheMu.Unlock()
		el.Value = en
		c.entries.MoveToFront(el)
	} else {
		var remEn *entry
		el = c.entries.PushFront(en)
		c.cache[en.key] = el
		if c.maximumSize > 0 && c.entries.Len() > c.maximumSize {
			remEn = c.removeOldest()
		}
		c.cacheMu.Unlock()
		if c.onInsertion != nil {
			c.onInsertion(en.key, en.value)
		}
		if c.onRemoval != nil && remEn != nil {
			c.onRemoval(remEn.key, remEn.value)
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
	if c.expireAfterAccess > 0 {
		getEntry(el).expire = currentTime().Add(c.expireAfterAccess)
	}
	c.entries.MoveToFront(el)
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
	now := currentTime()
	for i := drainMax; i > 0; i-- {
		el := c.entries.Back()
		if el == nil {
			// List is empty
			return
		}
		en := getEntry(el)
		if en.expire.IsZero() {
			// This should not happen
			return
		}
		if now.Before(en.expire) {
			// Expired
			return
		}
		c.remove(el)
	}
}

// removeOldest removes last element in entries list and returns removed entry.
// Calling this function must be guarded by entries and cache mutex.
func (c *localCache) removeOldest() *entry {
	el := c.entries.Back()
	if el == nil {
		return nil
	}
	en := getEntry(el)
	delete(c.cache, en.key)
	c.entries.Remove(el)
	c.stats.AddEvictionCount(1)
	return en
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

// withInsertionListener is used for testing.
func withInsertionListener(onInsertion Func) Option {
	return func(c *localCache) {
		c.onInsertion = onInsertion
	}
}
