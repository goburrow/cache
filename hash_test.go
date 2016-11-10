package cache

import (
	"encoding/binary"
	"hash/fnv"
	"testing"
	"unsafe"
)

func sumFNV(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

func sumFNVu64(v uint64) uint64 {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return sumFNV(b)
}

func TestSum(t *testing.T) {
	var tests = []struct {
		k interface{}
		h uint64
	}{
		{int(-1), sumFNVu64(^uint64(1) + 1)},
		{int8(-8), sumFNVu64(^uint64(8) + 1)},
		{int16(-16), sumFNVu64(^uint64(16) + 1)},
		{int32(-32), sumFNVu64(^uint64(32) + 1)},
		{int64(-64), sumFNVu64(^uint64(64) + 1)},
		{uint(1), sumFNVu64(1)},
		{uint8(8), sumFNVu64(8)},
		{uint16(16), sumFNVu64(16)},
		{uint32(32), sumFNVu64(32)},
		{uint64(64), sumFNVu64(64)},
		{byte(255), sumFNVu64(255)},
		{rune(1024), sumFNVu64(1024)},
		{true, sumFNVu64(1)},
		{false, sumFNVu64(0)},
		{float32(2.5), sumFNVu64(0x40200000)},
		{float64(2.5), sumFNVu64(0x4004000000000000)},
		{uintptr(unsafe.Pointer(t)), sumFNVu64(uint64(uintptr(unsafe.Pointer(t))))},
		{"", 0xcbf29ce484222325},
		{"string", sumFNV([]byte("string"))},
		{t, sumFNVu64(uint64(uintptr(unsafe.Pointer(t))))},
		{(*testing.T)(nil), sumFNVu64(0)},
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
