package cache

import (
	"errors"
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

	wg := sync.WaitGroup{}
	c := New(withInsertionListener(func(Key, Value) {
		wg.Done()
	}))
	defer c.Close()

	wg.Add(len(data))
	for _, d := range data {
		c.Put(d.k, d.v)
	}
	wg.Wait()

	for _, d := range data {
		v, ok := c.GetIfPresent(d.k)
		if !ok || v.(int) != d.v {
			t.Fatalf("unexpected value: %v (%v)", v, ok)
		}
	}
}

func TestMaximumSize(t *testing.T) {
	max := 10
	wg := sync.WaitGroup{}
	insFunc := func(k Key, v Value) {
		wg.Done()
	}
	c := New(WithMaximumSize(max), withInsertionListener(insFunc)).(*localCache)
	defer c.Close()

	wg.Add(max)
	for i := 0; i < max; i++ {
		c.Put(i, i)
	}
	wg.Wait()
	n := cacheSize(&c.cache)
	if n != max {
		t.Fatalf("unexpected cache size: %v, want: %v", n, max)
	}
	c.onInsertion = nil
	for i := 0; i < 2*max; i++ {
		k := rand.Intn(2 * max)
		c.Put(k, k)
		time.Sleep(time.Duration(i+1) * time.Millisecond)
		n = cacheSize(&c.cache)
		if n != max {
			t.Fatalf("unexpected cache size: %v, want: %v", n, max)
		}
	}
}

func TestRemovalListener(t *testing.T) {
	removed := make(map[Key]int)
	wg := sync.WaitGroup{}
	remFunc := func(k Key, v Value) {
		removed[k] = v.(int)
		wg.Done()
	}
	insFunc := func(k Key, v Value) {
		wg.Done()
	}
	max := 3
	c := New(WithMaximumSize(max), WithRemovalListener(remFunc),
		withInsertionListener(insFunc))
	defer c.Close()

	wg.Add(max + 2)
	for i := 1; i < max+2; i++ {
		c.Put(i, i)
	}
	wg.Wait()

	if len(removed) != 1 || removed[1] != 1 {
		t.Fatalf("unexpected removed entries: %+v", removed)
	}

	wg.Add(1)
	c.Invalidate(3)
	wg.Wait()
	if len(removed) != 2 || removed[3] != 3 {
		t.Fatalf("unexpected removed entries: %+v", removed)
	}
	wg.Add(2)
	c.InvalidateAll()
	wg.Wait()
	if len(removed) != 4 || removed[2] != 2 || removed[4] != 4 {
		t.Fatalf("unexpected removed entries: %+v", removed)
	}
}

func TestClose(t *testing.T) {
	removed := 0
	wg := sync.WaitGroup{}
	remFunc := func(Key, Value) {
		removed++
		wg.Done()
	}
	insFunc := func(Key, Value) {
		wg.Done()
	}
	c := New(WithRemovalListener(remFunc), withInsertionListener(insFunc))
	n := 10
	wg.Add(n)
	for i := 0; i < n; i++ {
		c.Put(i, i)
	}
	wg.Wait()
	wg.Add(n)
	c.Close()
	wg.Wait()
	if removed != n {
		t.Fatalf("unexpected removed: %d", removed)
	}
}

func TestLoadingCache(t *testing.T) {
	loadCount := 0
	loader := func(k Key) (Value, error) {
		loadCount++
		if k.(int)%2 != 0 {
			return nil, errors.New("odd")
		}
		return k, nil
	}
	wg := sync.WaitGroup{}
	insFunc := func(Key, Value) {
		wg.Done()
	}
	c := NewLoadingCache(loader, withInsertionListener(insFunc))
	defer c.Close()
	wg.Add(1)
	v, err := c.Get(2)
	if err != nil {
		t.Fatal(err)
	}
	if v.(int) != 2 {
		t.Fatalf("unexpected get: %v", v)
	}
	if loadCount != 1 {
		t.Fatalf("unexpected load count: %v", loadCount)
	}
	wg.Wait()
	v, err = c.Get(2)
	if err != nil {
		t.Fatal(err)
	}
	if v.(int) != 2 {
		t.Fatalf("unexpected get: %v", v)
	}
	if loadCount != 1 {
		t.Fatalf("unexpected load count: %v", loadCount)
	}
	v, err = c.Get(1)
	if err == nil || err.Error() != "odd" {
		t.Fatalf("expected error: %v", err)
	}
	// Should not insert
	wg.Wait()
}

func simpleLoader(k Key) (Value, error) {
	return k, nil
}

func TestCacheStats(t *testing.T) {
	wg := sync.WaitGroup{}
	insFunc := func(Key, Value) {
		wg.Done()
	}
	c := NewLoadingCache(simpleLoader, withInsertionListener(insFunc))
	defer c.Close()

	wg.Add(1)
	_, err := c.Get("x")
	if err != nil {
		t.Fatal(err)
	}
	var st Stats
	c.Stats(&st)
	if st.MissCount != 1 || st.LoadSuccessCount != 1 || st.TotalLoadTime <= 0 {
		t.Fatalf("unexpected stats: %+v", st)
	}
	wg.Wait()
	_, err = c.Get("x")
	if err != nil {
		t.Fatal(err)
	}
	c.Stats(&st)
	if st.HitCount != 1 {
		t.Fatalf("unexpected stats: %+v", st)
	}
}

func TestExpireAfterAccess(t *testing.T) {
	wg := sync.WaitGroup{}
	fn := func(k Key, v Value) {
		wg.Done()
	}
	mockTime := newMockTime()
	currentTime = mockTime.now
	c := New(WithExpireAfterAccess(1*time.Second), WithRemovalListener(fn),
		withInsertionListener(fn)).(*localCache)
	defer c.Close()

	wg.Add(1)
	c.Put(1, 1)
	wg.Wait()

	mockTime.add(1 * time.Second)
	wg.Add(2)
	c.Put(2, 2)
	c.Put(3, 3)
	wg.Wait()
	n := cacheSize(&c.cache)
	if n != 3 {
		wg.Add(n)
		t.Fatalf("unexpected cache size: %d, want: %d", n, 3)
	}

	mockTime.add(1 * time.Nanosecond)
	wg.Add(2)
	c.Put(4, 4)
	wg.Wait()
	n = cacheSize(&c.cache)
	wg.Add(n)
	if n != 3 {
		t.Fatalf("unexpected cache size: %d, want: %d", n, 3)
	}
	_, ok := c.GetIfPresent(1)
	if ok {
		t.Fatalf("unexpected entry status: %v, want: %v", ok, false)
	}
}

func TestExpireAfterWrite(t *testing.T) {
	loadCount := 0
	loader := func(k Key) (Value, error) {
		loadCount++
		return loadCount, nil
	}
	wg := sync.WaitGroup{}
	insFunc := func(Key, Value) {
		wg.Done()
	}
	mockTime := newMockTime()
	currentTime = mockTime.now
	c := NewLoadingCache(loader, WithExpireAfterWrite(1*time.Second),
		withInsertionListener(insFunc))
	defer c.Close()
	// New value
	wg.Add(1)
	v, err := c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if v.(int) != 1 || loadCount != 1 {
		t.Fatalf("unexpected load count: %v value=%v, want: %v value=%v", loadCount, v, 1, 1)
	}
	// Within 1s, the value should not yet expired.
	mockTime.add(1 * time.Second)
	v, err = c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	if v.(int) != 1 || loadCount != 1 {
		t.Fatalf("unexpected load count: %v value=%v, want: %v value=%v", loadCount, v, 1, 1)
	}
	// After 1s, the value should be expired and refresh triggered.
	mockTime.add(1 * time.Nanosecond)
	wg.Add(1)
	v, err = c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if v.(int) != 1 || loadCount != 2 {
		t.Fatalf("unexpected load count: %v value=%v, want: %v value=%v", loadCount, v, 2, 1)
	}
	// New value is loaded.
	v, err = c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	if v.(int) != 2 || loadCount != 2 {
		t.Fatalf("unexpected load count: %v value=%v, want: %v value=%v", loadCount, v, 2, 2)
	}
}

func TestGetIfPresentExpired(t *testing.T) {
	wg := sync.WaitGroup{}
	insFunc := func(Key, Value) {
		wg.Done()
	}
	c := New(WithExpireAfterWrite(1*time.Second), withInsertionListener(insFunc))
	mockTime := newMockTime()
	currentTime = mockTime.now

	v, ok := c.GetIfPresent(0)
	if ok {
		t.Fatalf("expect not present, actual: %v %v", v, ok)
	}
	wg.Add(1)
	c.Put(0, "0")
	v, ok = c.GetIfPresent(0)
	if !ok || v.(string) != "0" {
		t.Fatalf("expect present, actual: %v %v", v, ok)
	}
	wg.Wait()
	mockTime.add(2 * time.Second)
	v, ok = c.GetIfPresent(0)
	if ok {
		t.Fatalf("expect not present, actual: %v %v", v, ok)
	}
}

func TestSynchronousReload(t *testing.T) {
	var val Value
	loader := func(k Key) (Value, error) {
		time.Sleep(1 * time.Millisecond)
		return val, nil
	}
	executor := func(f func()) {
		f()
	}
	c := NewLoadingCache(loader, WithExpireAfterWrite(1*time.Second), WithExecutor(executor))
	val = "a"
	v, err := c.Get(1)
	if err != nil || v != val {
		t.Fatalf("unexpected get %v %v", v, err)
	}
	val = "b"
	v, err = c.Get(1)
	if err != nil || v != val {
		t.Fatalf("unexpected get %v %v", v, err)
	}
}

func TestCloseMultiple(t *testing.T) {
	c := New()
	start := make(chan bool)
	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			c.Close()
		}()
	}
	close(start)
	wg.Wait()
}

func BenchmarkGetSame(b *testing.B) {
	c := New()
	c.Put("*", "*")
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.GetIfPresent("*")
		}
	})
}

func cacheSize(c *cache) int {
	length := 0
	c.walk(func(*entry) {
		length++
	})
	return length
}

// mockTime is used for tests which required current system time.
type mockTime struct {
	mu    sync.RWMutex
	value time.Time
}

func newMockTime() *mockTime {
	return &mockTime{
		value: time.Now(),
	}
}

func (t *mockTime) add(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.value = t.value.Add(d)
}

func (t *mockTime) now() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.value
}
