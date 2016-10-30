package cache

import "container/list"

// lruCache is a LRU cache.
type lruCache struct {
	cache *cache
	ls    list.List
}

// init initializes cache list.
func (l *lruCache) init(c *cache) {
	l.cache = c
	l.ls.Init()
}

// add addes new entry to the cache and returns evicted entry if necessary.
func (l *lruCache) add(en *entry) *entry {
	l.cache.mu.Lock()
	defer l.cache.mu.Unlock()

	el := l.cache.data[en.key]
	if el != nil {
		// Entry had been added
		setEntry(el, en)
		l.ls.MoveToFront(el)
		return nil
	}
	if l.cache.cap <= 0 || l.ls.Len() < l.cache.cap {
		// Add this entry
		el = l.ls.PushFront(en)
		l.cache.data[en.key] = el
		return nil
	}
	// Replace with the last one
	el = l.ls.Back()
	if el == nil {
		// Can happen if cap is zero
		return en
	}
	remEn := getEntry(el)
	setEntry(el, en)
	l.ls.MoveToFront(el)

	delete(l.cache.data, remEn.key)
	l.cache.data[en.key] = el
	return remEn
}

// length returns length of the cache list.
func (l *lruCache) length() int {
	return l.ls.Len()
}

// access updates cache entry for a get.
func (l *lruCache) access(el *list.Element) {
	l.ls.MoveToFront(el)
}

// remove removes an entry from the cache.
func (l *lruCache) remove(el *list.Element) *entry {
	en := getEntry(el)
	l.cache.mu.Lock()
	defer l.cache.mu.Unlock()

	if _, ok := l.cache.data[en.key]; !ok {
		return nil
	}
	l.ls.Remove(el)
	delete(l.cache.data, en.key)
	return en
}
