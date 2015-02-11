package cachemap

import (
	"time"
)

////////////////////////////////////////////////////////////////////////////////
type Uint64CacheMap map[uint64]CacheObject

func (m Uint64CacheMap) Put(key uint64, obj CacheObject) {
	m[key] = obj
}
func (m Uint64CacheMap) Get(key uint64, now time.Time) (obj interface{}, ok bool) {
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
func (m Uint64CacheMap) Delete(key uint64) (ok bool) {
	delete(m, key)
	return true
}

type uint64CacheObjectWapper struct {
	obj CacheObject
	key uint64
}
type resultWapper struct {
	obj interface{}
	ok  bool
}
type resultGetter struct {
	key    uint64
	now    time.Time
	result chan resultWapper
}
type sizeGetter struct {
	size chan int
}
type Uint64SafeCacheMap struct {
	m                 Uint64CacheMap
	setChan           chan uint64CacheObjectWapper
	getChan           chan resultGetter
	delChan           chan uint64
	cleanerTimer      chan bool
	sizeChan          chan sizeGetter
	autoCleanInterval time.Duration
}

func NewUint64SafeCacheMap(autoCleanInterval time.Duration) (m *Uint64SafeCacheMap) {
	m = &Uint64SafeCacheMap{
		m:                 make(Uint64CacheMap),
		setChan:           make(chan uint64CacheObjectWapper),
		getChan:           make(chan resultGetter),
		delChan:           make(chan uint64),
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
func (safeMap *Uint64SafeCacheMap) Put(key uint64, obj CacheObject) {
	safeMap.setChan <- uint64CacheObjectWapper{key: key, obj: obj}
}
func (safeMap *Uint64SafeCacheMap) Get(key uint64, now time.Time) (obj interface{}, ok bool) {
	getter := resultGetter{key: key, now: now, result: make(chan resultWapper)}
	safeMap.getChan <- getter
	result := <-getter.result
	obj = result.obj
	ok = result.ok
	close(getter.result)
	return
}

func (safeMap *Uint64SafeCacheMap) Size() (size int) {
	sizeGetter := sizeGetter{size: make(chan int)}
	safeMap.sizeChan <- sizeGetter
	size = <-sizeGetter.size
	close(sizeGetter.size)
	return
}

func (safeMap *Uint64SafeCacheMap) Delete(key uint64) bool {
	safeMap.delChan <- key
	return true
}
func (safeMap *Uint64SafeCacheMap) GetDirtyKeys() (keys []uint64) {
	keys = make([]uint64, 0, len(safeMap.m))
	for key, _ := range safeMap.m {
		keys = append(keys, key)
	}
	return
}
