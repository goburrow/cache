package cache

import (
	"hash/fnv"
	"testing"
	"unsafe"
)

func sumFNV(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

func TestSum(t *testing.T) {
	var tests = []struct {
		k interface{}
		h uint64
	}{
		{int(-1), ^uint64(1) + 1},
		{int8(-8), ^uint64(8) + 1},
		{int16(-16), ^uint64(16) + 1},
		{int32(-32), ^uint64(32) + 1},
		{int64(-64), ^uint64(64) + 1},
		{uint(1), 1},
		{uint8(8), 8},
		{uint16(16), 16},
		{uint32(32), 32},
		{uint64(64), 64},
		{byte(255), 255},
		{rune(1024), 1024},
		{true, 1},
		{false, 0},
		{float32(2.5), 0x40200000},
		{float64(2.5), 0x4004000000000000},
		{uintptr(unsafe.Pointer(t)), uint64(uintptr(unsafe.Pointer(t)))},
		{"", 0xcbf29ce484222325},
		{"string", sumFNV([]byte("string"))},
		{t, uint64(uintptr(unsafe.Pointer(t)))},
		{(*testing.T)(nil), 0},
	}

	for _, tt := range tests {
		h := sum(tt.k)
		if h != tt.h {
			t.Errorf("unexpected hash: %v (0x%x), key: %v (%T), want: %v",
				h, h, tt.k, tt.k, tt.h)
		}
	}
}

func BenchmarkSumInt(b *testing.B) {
	var i interface{} = 0x0501
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sum(i)
		}
	})
}

func BenchmarkSumString(b *testing.B) {
	var s interface{} = "0913"
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sum(s)
		}
	})
}
