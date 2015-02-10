package cachemap

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCacheMap(t *testing.T) {
	m := make(Uint64CacheMap)
	now := time.Now()
	obj := NewCacheObject(100, now, 1)
	m.Put(1, obj)
	cacheObj, ok := m.Get(1, now)
	assert.True(t, ok)
	assert.True(t, 100 == cacheObj.(int))
	time.Sleep(1001 * time.Millisecond)
	cacheObj, ok = m.Get(1, time.Now())
	assert.False(t, ok)

	cacheObj, ok = m.Get(2, time.Now())
	assert.False(t, ok)

}

func TestSafeCacheMap(t *testing.T) {
	m := NewUint64SafeCacheMap(time.Microsecond)
	now := time.Now()
	obj := NewCacheObject(100, now, 1)
	m.Put(1, obj)
	cacheObj, ok := m.Get(1, now)
	assert.True(t, ok)
	assert.True(t, 100 == cacheObj.(int))
	assert.Equal(t, 1, m.Size())
	time.Sleep(1001 * time.Millisecond)
	assert.Equal(t, 0, m.Size()) // expires obj should be deleted
	cacheObj, ok = m.Get(1, time.Now())
	assert.False(t, ok)

	cacheObj, ok = m.Get(2, time.Now())
	assert.False(t, ok)

}

func TestSafeCacheMap32(t *testing.T) {
	m := NewInt32SafeCacheMap(time.Microsecond)
	now := time.Now()
	obj := NewCacheObject(100, now, 1)
	m.Put(1, obj)
	cacheObj, ok := m.Get(1, now)
	assert.True(t, ok)
	assert.True(t, 100 == cacheObj.(int))
	assert.Equal(t, 1, m.Size())
	time.Sleep(1001 * time.Millisecond)
	assert.Equal(t, 0, m.Size()) // expires obj should be deleted
	cacheObj, ok = m.Get(1, time.Now())
	assert.False(t, ok)

	cacheObj, ok = m.Get(2, time.Now())
	assert.False(t, ok)

}

func TestSafeCacheMapString(t *testing.T) {
	m := NewStringSafeCacheMap(time.Microsecond)
	now := time.Now()
	obj := NewCacheObject("hello", now, 1)
	m.Put("key", obj)
	cacheObj, ok := m.Get("key", now)
	assert.True(t, ok)
	assert.True(t, "hello" == cacheObj.(string))
	assert.Equal(t, 1, m.Size())
	time.Sleep(1001 * time.Millisecond)
	assert.Equal(t, 0, m.Size()) // expires obj should be deleted
	cacheObj, ok = m.Get("key", time.Now())
	assert.False(t, ok)

	cacheObj, ok = m.Get("key_not_exists", time.Now())
	assert.False(t, ok)

}
