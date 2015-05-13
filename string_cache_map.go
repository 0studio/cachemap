package cachemap

import (
	"time"
)

////////////////////////////////////////////////////////////////////////////////
type StringCacheMap map[string]CacheObject

func (m StringCacheMap) Put(key string, obj CacheObject) {
	m[key] = obj
}
func (m StringCacheMap) Get(key string, now time.Time) (obj interface{}, ok bool) {
	var cacheObj CacheObject
	cacheObj, ok = m[key]
	if !ok {
		return
	}
	obj, ok = cacheObj.GetObject(now)
	if !ok { // expired
		m.Delete(key)
	}
	return

}
func (m StringCacheMap) Delete(key string) (ok bool) {
	delete(m, key)
	return true
}

type stringCacheObjectWapper struct {
	obj CacheObject
	key string
}
type stringResultGetter struct {
	key    string
	now    time.Time
	result chan resultWapper
}
type StringSafeCacheMap struct {
	m                 StringCacheMap
	setChan           chan stringCacheObjectWapper
	getChan           chan stringResultGetter
	delChan           chan string
	cleanerTimer      chan bool
	sizeChan          chan sizeGetter
	autoCleanInterval time.Duration
}

func NewStringSafeCacheMap(autoCleanInterval time.Duration) (m *StringSafeCacheMap) {
	m = &StringSafeCacheMap{
		m:                 make(StringCacheMap),
		setChan:           make(chan stringCacheObjectWapper),
		getChan:           make(chan stringResultGetter),
		delChan:           make(chan string),
		cleanerTimer:      make(chan bool),
		sizeChan:          make(chan sizeGetter),
		autoCleanInterval: autoCleanInterval,
	}
	go func() {
		for {
			select {
			case setter := <-m.setChan:
				m.m.Put(setter.key, setter.obj)
			case getter := <-m.getChan:
				ret, ok := m.m.Get(getter.key, getter.now)
				go func() { getter.result <- resultWapper{obj: ret, ok: ok} }()
			case delId := <-m.delChan:
				m.m.Delete(delId)
			case <-m.cleanerTimer:
				now := time.Now()
				keys := m.GetDirtyKeys()
				for _, key := range keys {
					m.m.Get(key, now) // m.Get will delete expire obj
				}
			case sizeGetter := <-m.sizeChan:
				go func() { sizeGetter.size <- len(m.m) }()
			}
		}
	}()
	go func() {
		for {
			select {
			case <-time.After(m.autoCleanInterval):
				m.cleanerTimer <- true
			}
		}
	}()
	return
}
func (safeMap *StringSafeCacheMap) Put(key string, obj CacheObject) {
	safeMap.setChan <- stringCacheObjectWapper{key: key, obj: obj}
}
func (safeMap *StringSafeCacheMap) Get(key string, now time.Time) (obj interface{}, ok bool) {
	getter := stringResultGetter{key: key, now: now, result: make(chan resultWapper)}
	safeMap.getChan <- getter
	result := <-getter.result
	obj = result.obj
	ok = result.ok
	close(getter.result)
	return
}

func (safeMap *StringSafeCacheMap) Size() (size int) {
	sizeGetter := sizeGetter{size: make(chan int)}
	safeMap.sizeChan <- sizeGetter
	size = <-sizeGetter.size
	close(sizeGetter.size)
	return
}

func (safeMap *StringSafeCacheMap) Delete(key string) bool {
	safeMap.delChan <- key
	return true
}
func (safeMap *StringSafeCacheMap) GetDirtyKeys() (keys []string) {
	keys = make([]string, 0, len(safeMap.m))
	for key, _ := range safeMap.m {
		keys = append(keys, key)
	}
	return
}
