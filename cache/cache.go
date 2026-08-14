package cache

import (
	"sync"
	"time"
)

// Entry represents a cached value with an expiration time.
type Entry struct {
	Value     string
	Content   interface{}
	ExpiresAt time.Time
}

// Cache is a thread-safe in-memory cache with TTL support.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]Entry
	done    chan struct{}
}

// New creates a new Cache and starts a background cleanup goroutine.
func New() *Cache {
	c := &Cache{
		entries: make(map[string]Entry),
		done:    make(chan struct{}),
	}
	go c.cleanup()
	return c
}

// Close stops the background cleanup goroutine.
func (c *Cache) Close() {
	close(c.done)
}

// Get retrieves a value from the cache. Returns the value and true if found and not expired.
func (c *Cache) Get(key string) (Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok {
		return Entry{}, false
	}
	if time.Now().After(entry.ExpiresAt) {
		return Entry{}, false
	}
	return entry, true
}

// Set stores a value in the cache with the given TTL.
func (c *Cache) Set(key string, value string, content interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = Entry{
		Value:     value,
		Content:   content,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// cleanup periodically removes expired entries every 10 minutes.
func (c *Cache) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.entries {
				if now.After(v.ExpiresAt) {
					delete(c.entries, k)
				}
			}
			c.mu.Unlock()
		}
	}
}
