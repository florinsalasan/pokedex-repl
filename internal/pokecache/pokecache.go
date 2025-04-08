package pokecache

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
    entries map[string]cacheEntry
    mu sync.Mutex
    interval time.Duration
}

type cacheEntry struct {
    createdAt time.Time
    val []byte
}

func NewCache(interval time.Duration) *Cache {
    pokecache := Cache{entries: make(map[string]cacheEntry), interval: interval}
    go pokecache.reapLoop()

    return &pokecache
}

func (c *Cache) Add(key string, val []byte) {
    // fmt.Println("adding key-val to cache")
    newEntry := cacheEntry{val: val, createdAt: time.Now()}
    c.mu.Lock()
    defer c.mu.Unlock()
    c.entries[key] = newEntry
}

func (c *Cache) Get(key string) ([]byte, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.entries[key].val != nil {
        return c.entries[key].val, true
    }
    return nil, false
}

func (c *Cache) reapLoop() {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        <-ticker.C
        currTime := time.Now()

        c.mu.Lock()

        for key, value := range c.entries {
            if currTime.Sub(value.createdAt) > c.interval {
                delete(c.entries, key)
            }
        }

        c.mu.Unlock()
    }
}
