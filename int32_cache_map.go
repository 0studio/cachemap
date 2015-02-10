package cachemap

import (
	"time"
)

////////////////////////////////////////////////////////////////////////////////
type Int32CacheMap map[uint32]CacheObject

func (m Int32CacheMap) Put(key uint32, obj CacheObject) {
	m[key] = obj
}
func (m Int32CacheMap) Get(key uint32, now time.Time) (obj interface{}, ok bool) {
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
func (m Int32CacheMap) Delete(key uint32) (ok bool) {
	delete(m, key)
	return true
}

type uint32CacheObjectWapper struct {
	obj CacheObject
	key uint32
}
type int32resultGetter struct {
	key    uint32
	now    time.Time
	result chan resultWapper
}
type Int32SafeCacheMap struct {
	m                 Int32CacheMap
	setChan           chan uint32CacheObjectWapper
	getChan           chan int32resultGetter
	delChan           chan uint32
	cleanerTimer      chan bool
	sizeChan          chan sizeGetter
	autoCleanInterval time.Duration
}

func NewInt32SafeCacheMap(autoCleanInterval time.Duration) (m *Int32SafeCacheMap) {
	m = &Int32SafeCacheMap{
		m:                 make(Int32CacheMap),
		setChan:           make(chan uint32CacheObjectWapper),
		getChan:           make(chan int32resultGetter),
		delChan:           make(chan uint32),
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
				for key, _ := range m.m {
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
func (safeMap *Int32SafeCacheMap) Put(key uint32, obj CacheObject) {
	safeMap.setChan <- uint32CacheObjectWapper{key: key, obj: obj}
}
func (safeMap *Int32SafeCacheMap) Get(key uint32, now time.Time) (obj interface{}, ok bool) {
	getter := int32resultGetter{key: key, now: now, result: make(chan resultWapper)}
	safeMap.getChan <- getter
	result := <-getter.result
	obj = result.obj
	ok = result.ok
	close(getter.result)
	return
}

func (safeMap *Int32SafeCacheMap) Size() (size int) {
	sizeGetter := sizeGetter{size: make(chan int)}
	safeMap.sizeChan <- sizeGetter
	size = <-sizeGetter.size
	close(sizeGetter.size)
	return
}

func (safeMap *Int32SafeCacheMap) Delete(key uint32) bool {
	safeMap.delChan <- key
	return true
}
