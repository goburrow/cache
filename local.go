package cache

import (
	"container/list"
	"sync"
	"time"
)

const (
	defaultMaximumSize = 1<<31 - 1
	defaultChanBufSize = 64
)

// currentTime is an alias for time.Now, used for testing.
var currentTime = time.Now

// entry stores cached entry key and value.
type entry struct {
	key   Key
	value Value

	// accessed is the last accessed time
	accessed time.Time
}

func getEntry(el *list.Element) *entry {
	return el.Value.(*entry)
}

// localCache is an asynchronous LRU cache.
type localCache struct {
	maximumSize int

	onInsertion Func
	onRemoval   Func

	loader LoaderFunc

	cacheMu sync.RWMutex
	cache   map[Key]*list.Element

	entries     list.List
	addEntry    chan *entry
	accessEntry chan *list.Element
	deleteEntry chan *list.Element

	closeCh chan struct{}
}

// newLocalCache returns a default localCache
func newLocalCache() *localCache {
	c := &localCache{
		maximumSize: defaultMaximumSize,
		cache:       make(map[Key]*list.Element),

		addEntry:    make(chan *entry, defaultChanBufSize),
		accessEntry: make(chan *list.Element, defaultChanBufSize),
		deleteEntry: make(chan *list.Element, defaultChanBufSize),
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
		v := getEntry(el).value
		c.accessEntry <- el
		return v, true
	}
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
		v := getEntry(el).value
		c.accessEntry <- el
		return v, nil
	}
	if c.loader == nil {
		panic("loader must be set")
	}
	v, err := c.loader(k)
	if err != nil {
		return nil, err
	}
	en := &entry{
		key:   k,
		value: v,
	}
	c.addEntry <- en
	return v, nil
}

func (c *localCache) processEntries() {
	for {
		select {
		case <-c.closeCh:
			c.removeAll()
			return
		case en := <-c.addEntry:
			en.accessed = currentTime()
			c.add(en)
		case el := <-c.accessEntry:
			getEntry(el).accessed = currentTime()
			c.entries.MoveToFront(el)
		case el := <-c.deleteEntry:
			if el == nil {
				c.removeAll()
			} else {
				c.remove(el)
			}
		}
	}
}

func (c *localCache) add(en *entry) {
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

// withInsertionListener is used for testing.
func withInsertionListener(onInsertion Func) Option {
	return func(c *localCache) {
		c.onInsertion = onInsertion
	}
}
