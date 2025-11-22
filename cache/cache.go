package cache

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration time
type CacheItem struct {
	Value      any
	Expiration int64
}

// IsExpired checks if the cache item has expired
func (item CacheItem) IsExpired() bool {
	return time.Now().UnixNano() > item.Expiration
}

// Cache represents an in-memory cache with TTL
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
	ttl   time.Duration
}

// New creates a new cache instance with the specified TTL
func New(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]CacheItem),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Set stores a value in the cache with TTL
func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(c.ttl).UnixNano()
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if item.IsExpired() {
		delete(c.items, key)
		return nil, false
	}

	return item.Value, true
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]CacheItem)
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// cleanup removes expired items from the cache periodically
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute) // Clean up every minute
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now().UnixNano()
		for key, item := range c.items {
			if now > item.Expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Keys returns all keys in the cache (for debugging)
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}
