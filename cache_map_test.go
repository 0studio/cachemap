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
