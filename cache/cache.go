package cache

import (
	"container/list"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration time
type CacheItem struct {
	Key        string
	Value      any
	Size       int64
	Expiration int64
}

// IsExpired checks if the cache item has expired
func (item *CacheItem) IsExpired() bool {
	return time.Now().UnixNano() > item.Expiration
}

// Cache represents an in-memory LRU cache with TTL
type Cache struct {
	items       map[string]*list.Element
	evictList   *list.List
	mu          sync.RWMutex
	ttl         time.Duration
	maxSize     int64
	currentSize int64
}

// New creates a new cache instance with the specified TTL and max size (in bytes)
func New(ttl time.Duration, maxSize int64) *Cache {
	c := &Cache{
		items:     make(map[string]*list.Element),
		evictList: list.New(),
		ttl:       ttl,
		maxSize:   maxSize,
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Set stores a value in the cache with TTL and size
func (c *Cache) Set(key string, value any, size int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If item size is larger than max size, don't cache it
	if c.maxSize > 0 && size > c.maxSize {
		return
	}

	expiration := time.Now().Add(c.ttl).UnixNano()

	// Check if it already exists
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		item := ent.Value.(*CacheItem)
		c.currentSize -= item.Size
		c.currentSize += size
		item.Value = value
		item.Size = size
		item.Expiration = expiration
	} else {
		ent := c.evictList.PushFront(&CacheItem{
			Key:        key,
			Value:      value,
			Size:       size,
			Expiration: expiration,
		})
		c.items[key] = ent
		c.currentSize += size
	}

	// Evict older items if we exceed max size
	if c.maxSize > 0 {
		for c.currentSize > c.maxSize {
			ent := c.evictList.Back()
			if ent != nil {
				c.removeElement(ent)
			}
		}
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (any, bool) {
	c.mu.Lock() // Write lock needed for MoveToFront
	defer c.mu.Unlock()

	if ent, ok := c.items[key]; ok {
		item := ent.Value.(*CacheItem)
		if item.IsExpired() {
			c.removeElement(ent)
			return nil, false
		}
		c.evictList.MoveToFront(ent)
		return item.Value, true
	}

	return nil, false
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
	}
}

// removeElement removes a given list element from the cache.
// Caller must hold c.mu.
func (c *Cache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*CacheItem)
	delete(c.items, kv.Key)
	c.currentSize -= kv.Size
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.evictList = list.New()
	c.currentSize = 0
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
		// Iterating over the map while deleting is safe in Go
		for _, ent := range c.items {
			item := ent.Value.(*CacheItem)
			if now > item.Expiration {
				c.removeElement(ent)
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
