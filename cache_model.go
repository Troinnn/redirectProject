package main

import (
	"sync"
	"time"
)

type CachedLink struct {
	Link
	expireAtTimestamp int64 // время протухания кэша
}

//канал структуры кеш
type LocalCache struct {
	stop chan struct{} // канал структуры исп. для остановки

	wg    sync.WaitGroup
	mu    sync.RWMutex
	links map[string]CachedLink //наш кэш
}

type Cache interface {
	Add(key string, value string)
	Get(key string) (string, bool)
	Len() int
}

func newLocalCache(cleanupInterval time.Duration) *LocalCache {
	lc := &LocalCache{
		links: make(map[string]CachedLink),
		stop:  make(chan struct{}),
	}

	lc.wg.Add(1)
	go func(cleanupInterval time.Duration) {
		defer lc.wg.Done()
		lc.cleanupLoop(cleanupInterval)
	}(cleanupInterval)

	return lc
}

func (lc *LocalCache) cleanupLoop(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-lc.stop:
			return
		case <-t.C:
			lc.mu.Lock()
			for k, v := range lc.links {
				if v.expireAtTimestamp <= time.Now().UTC().Unix() {
					delete(lc.links, k)
				}
			}
			lc.mu.Unlock()
		}
	}
}

func (lc *LocalCache) Add(key string, value string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.links[key] = CachedLink{
		Link: Link{
			ID:          "",
			ActiveLink:  value,
			HistoryLink: "",
		},
		expireAtTimestamp: time.Now().UTC().Add(360 * time.Minute).Unix(),
	}
}

func (lc *LocalCache) Get(key string) (string, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	cu, ok := lc.links[key]
	if !ok {
		return "", false
	}

	return cu.Link.ActiveLink, true
}

func (lc *LocalCache) Len() int {
	return len(lc.links)
}
