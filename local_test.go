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
	if len(c.cache.data) != max {
		t.Fatalf("unexpected cache size: %v, want: %v", len(c.cache.data), max)
	}
	c.onInsertion = nil
	for i := 0; i < 2*max; i++ {
		k := rand.Intn(2 * max)
		c.Put(k, k)
		time.Sleep(time.Duration(i+1) * time.Millisecond)
		if len(c.cache.data) != max {
			t.Fatalf("unexpected cache size: %v, want: %v", len(c.cache.data), max)
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
	insFunc := func(Key, Value) {
		wg.Done()
	}
	now := time.Now()
	currentTime = func() time.Time {
		return now
	}
	c := NewLoadingCache(simpleLoader, WithExpireAfterAccess(1*time.Second),
		withInsertionListener(insFunc)).(*localCache)
	defer c.Close()

	wg.Add(1)
	_, err := c.Get("x")
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	currentTime = func() time.Time {
		return now.Add(1 * time.Second)
	}
	c.expireEntries()
	if len(c.cache.data) != 1 {
		t.Fatalf("unexpected cache size: %v, want: %v", len(c.cache.data), 1)
	}
	currentTime = func() time.Time {
		return now.Add(1*time.Second + 1)
	}
	c.expireEntries()
	if len(c.cache.data) != 0 {
		t.Fatalf("unexpected cache size: %v, want: %v", len(c.cache.data), 0)
	}
}

func TestRefreshAfterWrite(t *testing.T) {
	loadCount := 0
	loader := func(k Key) (Value, error) {
		loadCount++
		return loadCount, nil
	}
	wg := sync.WaitGroup{}
	insFunc := func(Key, Value) {
		wg.Done()
	}
	now := time.Now()
	currentTime = func() time.Time {
		return now
	}
	c := NewLoadingCache(loader, WithRefreshAfterWrite(1*time.Second),
		withInsertionListener(insFunc))
	defer c.Close()

	wg.Add(1)
	v, err := c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if v.(int) != 1 || loadCount != 1 {
		t.Fatalf("unexpected load count: %v, %v", v, loadCount)
	}

	currentTime = func() time.Time {
		return now.Add(1 * time.Second)
	}
	v, err = c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	if v.(int) != 1 || loadCount != 1 {
		t.Fatalf("unexpected load count: %v, %v", v, loadCount)
	}

	currentTime = func() time.Time {
		return now.Add(1*time.Second + 1)
	}
	wg.Add(1)
	v, err = c.Get("refresh")
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if v.(int) != 2 || loadCount != 2 {
		t.Fatalf("unexpected load count: %v, %v", v, loadCount)
	}
}
