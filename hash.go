package cache

import (
	"math"
	"reflect"
)

// Hash is an interface implemented by cache keys to
// override default hash function.
type Hash interface {
	Sum64() uint64
}

// sum calculates hash value of the given key.
func sum(k interface{}) uint64 {
	if h, ok := k.(Hash); ok {
		return h.Sum64()
	}
	switch h := k.(type) {
	case int:
		return hashU64(uint64(h))
	case int8:
		return hashU64(uint64(h))
	case int16:
		return hashU64(uint64(h))
	case int32:
		return hashU64(uint64(h))
	case int64:
		return hashU64(uint64(h))
	case uint:
		return hashU64(uint64(h))
	case uint8:
		return hashU64(uint64(h))
	case uint16:
		return hashU64(uint64(h))
	case uint32:
		return hashU64(uint64(h))
	case uint64:
		return hashU64(h)
	case uintptr:
		return hashU64(uint64(h))
	case float32:
		return hashU64(uint64(math.Float32bits(h)))
	case float64:
		return hashU64(math.Float64bits(h))
	case bool:
		if h {
			return hashU64(1)
		}
		return hashU64(0)
	case string:
		return hashBytes([]byte(h))
	}
	// TODO: complex64 and complex128
	if h, ok := hashPointer(k); ok {
		return h
	}
	// TODO: use gob to encode k to bytes then hash.
	return 0
}

const (
	fnvOffset = 14695981039346656037
	fnvPrime  = 1099511628211
)

func hashU64(v uint64) uint64 {
	// Inline code from hash/fnv to reduce memory allocations
	var h uint64 = fnvOffset
	for i := uint(0); i < 64; i += 8 {
		h ^= (v >> i) & 0xFF
		h *= fnvPrime
	}
	return h
}

// hashBytes calculates hash value using FNV-1a algorithm.
func hashBytes(data []byte) uint64 {
	// Inline code from hash/fnv to reduce memory allocations
	var h uint64 = fnvOffset
	for _, b := range data {
		h ^= uint64(b)
		h *= fnvPrime
	}
	return h
}

func hashPointer(k interface{}) (uint64, bool) {
	v := reflect.ValueOf(k)
	switch v.Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Func, reflect.Slice, reflect.Map, reflect.Chan:
		return hashU64(uint64(v.Pointer())), true
	}
	return 0, false
}
