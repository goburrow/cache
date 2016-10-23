// Package cache provides partial implementations of Guava Cache.
package cache

// Key is any value which is comparable.
// See http://golang.org/ref/spec#Comparison_operators for details.
type Key interface{}

// Value is any value.
type Value interface{}

// Cache is a key-value cache which entries are added and stayed in the
// cache until either are evicted or manually invalidated.
type Cache interface {
	// GetIfPresent returns value associated with Key or (nil, false)
	// if there is no cached value for Key.
	GetIfPresent(Key) (Value, bool)

	// Put associates value with Key. If a value is already associated
	// with Key, the old one will be replaced with Value.
	Put(Key, Value)

	// Invalidate discards cached value for Key
	Invalidate(Key)

	// InvalidateAll discards all entries.
	InvalidateAll()
}

// OnRemoval is a callback when an entry is evicted from cache.
type OnRemoval func(Key, Value)
