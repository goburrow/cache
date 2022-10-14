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
	c := New[string, int](withInsertionListener(func(string, int) {
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
		if !ok || v != d.v {
			t.Fatalf("unexpected value: %v (%v)", v, ok)
		}
	}
}

func TestMaximumSize(t *testing.T) {
	max := 10
	wg := sync.WaitGroup{}
	insFunc := func(k int, v int) {
		wg.Done()
	}
	c := New[int, int](WithMaximumSize[int, int](max), withInsertionListener(insFunc)).(*localCache[int, int])
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
	removed := make(map[int]int)
	wg := sync.WaitGroup{}
	remFunc := func(k int, v int) {
		removed[k] = v
		wg.Done()
	}
	insFunc := func(k int, v int) {
		wg.Done()
	}
	max := 3
	c := New(WithMaximumSize[int, int](max), WithRemovalListener(remFunc),
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
	remFunc := func(int, int) {
		removed++
		wg.Done()
	}
	insFunc := func(int, int) {
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
	loader := func(k int) (int, error) {
		loadCount++
		if k%2 != 0 {
			return 0, errors.New("odd")
		}
		return k, nil
	}
	wg := sync.WaitGroup{}
	insFunc := func(int, int) {
		wg.Done()
	}
	c := NewLoadingCache(loader, withInsertionListener(insFunc))
	defer c.Close()
	wg.Add(1)
	v, err := c.Get(2)
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 {
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
	if v != 2 {
		t.Fatalf("unexpected get: %v", v)
	}
	if loadCount != 1 {
		t.Fatalf("unexpected load count: %v", loadCount)
	}
	_, err = c.Get(1)
	if err == nil || err.Error() != "odd" {
		t.Fatalf("expected error: %v", err)
	}
	// Should not insert
	wg.Wait()
}

func simpleLoader(k string) (string, error) {
	return k, nil
}

func TestCacheStats(t *testing.T) {
	wg := sync.WaitGroup{}
	insFunc := func(string, string) {
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
	fn := func(k int, v int) {
		wg.Done()
	}
	mockTime := newMockTime()
	currentTime = mockTime.now
	c := New(WithExpireAfterAccess[int, int](1*time.Second), WithRemovalListener(fn),
		withInsertionListener(fn)).(*localCache[int, int])
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
	loader := func(k string) (int, error) {
		loadCount++
		return loadCount, nil
	}
	wg := sync.WaitGroup{}
	insFunc := func(string, int) {
		wg.Done()
	}
	mockTime := newMockTime()
	currentTime = mockTime.now
	c := NewLoadingCache(loader, WithExpireAfterWrite[string, int](1*time.Second),
		withInsertionListener(insFunc))
	defer c.Close()
	// New value
	wg.Add(1)
	v, err := c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if v != 1 || loadCount != 1 {
		t.Fatalf("unexpected load count: %v value=%v, want: %v value=%v", loadCount, v, 1, 1)
	}
	// Within 1s, the value should not yet expired.
	mockTime.add(1 * time.Second)
	v, err = c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 || loadCount != 1 {
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
	if v != 1 || loadCount != 2 {
		t.Fatalf("unexpected load count: %v value=%v, want: %v value=%v", loadCount, v, 2, 1)
	}
	// New value is loaded.
	v, err = c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 || loadCount != 2 {
		t.Fatalf("unexpected load count: %v value=%v, want: %v value=%v", loadCount, v, 2, 2)
	}
}

func TestRefreshAterWrite(t *testing.T) {
	var mutex sync.Mutex
	loaded := make(map[int]int)
	loader := func(k int) (int, error) {
		mutex.Lock()
		n := loaded[k]
		n++
		loaded[k] = n
		mutex.Unlock()
		return n, nil
	}
	wg := sync.WaitGroup{}
	insFunc := func(int, int) {
		wg.Done()
	}
	mockTime := newMockTime()
	currentTime = mockTime.now
	c := NewLoadingCache(loader, WithExpireAfterAccess[int, int](4*time.Second), WithRefreshAfterWrite[int, int](2*time.Second),
		WithReloader[int, int](&syncReloader[int, int]{loader}), withInsertionListener(insFunc))
	defer c.Close()

	wg.Add(3)
	v, err := c.Get(1)
	if err != nil || v != 1 {
		t.Fatalf("unexpected get: %v %v", v, err)
	}
	// 3s
	mockTime.add(3 * time.Second)
	v, err = c.Get(2)
	if err != nil || v != 1 {
		t.Fatalf("unexpected get: %v %v", v, err)
	}
	wg.Wait()
	if loaded[1] != 2 || loaded[2] != 1 {
		t.Fatalf("unexpected loaded: %v", loaded)
	}
	v, err = c.Get(1)
	if err != nil || v != 2 {
		t.Fatalf("unexpected get: %v %v", v, err)
	}
	// 8s
	mockTime.add(5 * time.Second)
	wg.Add(1)
	v, err = c.Get(1)
	if err != nil || v != 3 {
		t.Fatalf("unexpected get: %v %v", v, err)
	}
}

func TestGetIfPresentExpired(t *testing.T) {
	wg := sync.WaitGroup{}
	insFunc := func(int, string) {
		wg.Done()
	}
	c := New(WithExpireAfterWrite[int, string](1*time.Second), withInsertionListener(insFunc))
	mockTime := newMockTime()
	currentTime = mockTime.now

	v, ok := c.GetIfPresent(0)
	if ok {
		t.Fatalf("expect not present, actual: %v %v", v, ok)
	}
	wg.Add(1)
	c.Put(0, "0")
	v, ok = c.GetIfPresent(0)
	if !ok || v != "0" {
		t.Fatalf("expect present, actual: %v %v", v, ok)
	}
	wg.Wait()
	mockTime.add(2 * time.Second)
	v, ok = c.GetIfPresent(0)
	if ok {
		t.Fatalf("expect not present, actual: %v %v", v, ok)
	}
}

func TestLoadingAsyncReload(t *testing.T) {
	var val string
	loader := func(k int) (string, error) {
		if val == "" {
			return "", errors.New("empty")
		}
		return val, nil
	}
	mockTime := newMockTime()
	currentTime = mockTime.now
	c := NewLoadingCache(loader, WithExpireAfterWrite[int, string](5*time.Millisecond),
		WithReloader[int, string](&syncReloader[int, string]{loader}))
	val = "a"
	v, err := c.Get(1)
	if err != nil || v != val {
		t.Fatalf("unexpected get %v %v", v, err)
	}
	mockTime.add(50 * time.Millisecond)
	val = "b"
	v, err = c.Get(1)
	if err != nil || v != val {
		t.Fatalf("unexpected get %v %v", v, err)
	}
	val = ""
	v, err = c.Get(2)
	if v != "" || err == nil || err.Error() != "empty" {
		t.Fatalf("expect error: actual %v %v", v, err)
	}
}

func TestLoadingRefresh(t *testing.T) {
	count := 0
	c := NewLoadingCache(func(key int) (int, error) {
		count++
		return count, nil
	})
	for i := 10; i > 0; i-- {
		v, _ := c.Get(1)
		if v != 1 {
			t.Fatalf("expect value loaded, actual: %v", v)
		}
		v, ok := c.GetIfPresent(1)
		if !ok || v != 1 {
			t.Fatalf("expect value present, actual: %v %v", v, ok)
		}
	}
	c.Refresh(2)
	v, _ := c.Get(2)
	if v != 2 {
		t.Fatalf("expect new value loaded, actual: %v", v)
	}
	c.Put(2, 3)
	v, _ = c.Get(2)
	if v != 3 {
		t.Fatalf("expect new value, actual: %v", v)
	}
}

func TestCloseMultiple(t *testing.T) {
	c := New[int, int]()
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
	// Should not panic
	c.GetIfPresent(0)
	c.Put(1, 1)
	c.Invalidate(0)
	c.InvalidateAll()
	c.Close()
}

func BenchmarkGetSame(b *testing.B) {
	c := New[string, string]()
	c.Put("*", "*")
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.GetIfPresent("*")
		}
	})
}

func cacheSize[Key comparable, Value any](c *cache[Key, Value]) int {
	length := 0
	c.walk(func(*entry[Key, Value]) {
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

type syncReloader[Key comparable, Value any] struct {
	loaderFn LoaderFunc[Key, Value]
}

func (s *syncReloader[Key, Value]) Reload(k Key, v Value, setFn func(Value, error)) {
	v, err := s.loaderFn(k)
	setFn(v, err)
}

func (s *syncReloader[Key, Value]) Close() error {
	return nil
}
