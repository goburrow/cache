package cache

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	data := []struct {
		k string
		v int
	}{
		{"1", 1},
		{"2", 2},
	}

	c := New()
	for _, d := range data {
		c.Put(d.k, d.v)
	}

	for _, d := range data {
		v, ok := c.GetIfPresent(d.k)
		if !ok || v.(int) != d.v {
			t.Fatalf("unexpected value: %v (%v)", v, ok)
		}
	}
}

func TestCacheMaximumSize(t *testing.T) {
	max := 5
	c := New(WithMaximumSize(5)).(*localCache)

	for i := 0; i < max; i++ {
		c.Put(i, i)
	}
	for i := 0; i < 2*max; i++ {
		k := rand.Intn(2 * max)
		c.Put(k, k)
		if len(c.cache) != max || c.entries.Len() != max {
			t.Fatalf("unexpected cache size: %v, %v", len(c.cache), c.entries.Len())
		}
	}
}

func TestCacheConcurrency(t *testing.T) {
	max := 128
	c := New()

	wg := sync.WaitGroup{}
	wg.Add(max)
	for i := 0; i < max; i++ {
		go func(i int) {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(10))*time.Millisecond + 1)
			k := rand.Int63()
			c.Put(k, k)
			v, ok := c.GetIfPresent(k)
			if !ok || v.(int64) != k {
				t.Errorf("unexpected get: %v (%v)", v, ok)
			}
		}(i)
	}
	wg.Wait()
}

func TestRemovalListener(t *testing.T) {
	removed := make(map[Key]int)
	cb := func(k Key, v Value) {
		removed[k] = v.(int)
	}
	c := New(WithMaximumSize(3), WithRemovalListener(cb))
	c.Put(1, 1)
	c.Put(2, 2)
	c.Put(3, 3)
	c.Put(4, 4)
	if len(removed) != 1 || removed[1] != 1 {
		t.Fatalf("unexpected removed entries: %+v", removed)
	}

	c.Invalidate(3)
	if len(removed) != 2 || removed[3] != 3 {
		t.Fatalf("unexpected removed entries: %+v", removed)
	}
	c.InvalidateAll()
	if len(removed) != 4 || removed[2] != 2 || removed[4] != 4 {
		t.Fatalf("unexpected removed entries: %+v", removed)
	}
}

func BenchmarkCache(b *testing.B) {
	c := New(WithMaximumSize(1024))
	rand.Seed(time.Now().UnixNano())

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := rand.Int63()
			c.Put(k, b)
			k = rand.Int63()
			_, _ = c.GetIfPresent(k)
		}
	})
}
