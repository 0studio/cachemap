package cachemap

import (
	"time"
)

////////////////////////////////////////////////////////////////////////////////
type Uint32CacheMap map[uint32]CacheObject

func (m Uint32CacheMap) Put(key uint32, obj CacheObject) {
	m[key] = obj
}
func (m Uint32CacheMap) Get(key uint32, now time.Time) (obj interface{}, ok bool) {
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
func (m Uint32CacheMap) Delete(key uint32) (ok bool) {
	delete(m, key)
	return true
}

type uint32CacheObjectWapper struct {
	obj CacheObject
	key uint32
}
type uint32resultGetter struct {
	key    uint32
	now    time.Time
	result chan resultWapper
}
type Uint32SafeCacheMap struct {
	m                 Uint32CacheMap
	setChan           chan uint32CacheObjectWapper
	getChan           chan uint32resultGetter
	delChan           chan uint32
	cleanerTimer      chan bool
	sizeChan          chan sizeGetter
	autoCleanInterval time.Duration
}

func NewUint32SafeCacheMap(autoCleanInterval time.Duration) (m *Uint32SafeCacheMap) {
	m = &Uint32SafeCacheMap{
		m:                 make(Uint32CacheMap),
		setChan:           make(chan uint32CacheObjectWapper),
		getChan:           make(chan uint32resultGetter),
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
func (safeMap *Uint32SafeCacheMap) Put(key uint32, obj CacheObject) {
	safeMap.setChan <- uint32CacheObjectWapper{key: key, obj: obj}
}
func (safeMap *Uint32SafeCacheMap) Get(key uint32, now time.Time) (obj interface{}, ok bool) {
	getter := uint32resultGetter{key: key, now: now, result: make(chan resultWapper)}
	safeMap.getChan <- getter
	result := <-getter.result
	obj = result.obj
	ok = result.ok
	close(getter.result)
	return
}

func (safeMap *Uint32SafeCacheMap) Size() (size int) {
	sizeGetter := sizeGetter{size: make(chan int)}
	safeMap.sizeChan <- sizeGetter
	size = <-sizeGetter.size
	close(sizeGetter.size)
	return
}

func (safeMap *Uint32SafeCacheMap) Delete(key uint32) bool {
	safeMap.delChan <- key
	return true
}
func (safeMap *Uint32SafeCacheMap) GetDirtyKeys() (keys []uint32) {
	keys = make([]uint32, 0, len(safeMap.m))
	for key, _ := range safeMap.m {
		keys = append(keys, key)
	}
	return
}
