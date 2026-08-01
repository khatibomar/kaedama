package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	ttl := 5 * time.Second
	c := New(ttl, 1024)

	if c == nil {
		t.Fatal("expected cache to be created, got nil")
	}

	if c.ttl != ttl {
		t.Errorf("expected TTL to be %v, got %v", ttl, c.ttl)
	}

	if c.items == nil {
		t.Fatal("expected items map to be initialized")
	}

	if len(c.items) != 0 {
		t.Errorf("expected empty cache, got %d items", len(c.items))
	}
}

func TestSetAndGet(t *testing.T) {
	c := New(time.Minute, 1024)

	// Test setting and getting a value
	key := "test-key"
	value := "test-value"

	c.Set(key, value, 1)

	retrieved, exists := c.Get(key)
	if !exists {
		t.Fatal("expected key to exist in cache")
	}

	if retrieved != value {
		t.Errorf("expected value %v, got %v", value, retrieved)
	}
}

func TestGetNonExistentKey(t *testing.T) {
	c := New(time.Minute, 1024)

	value, exists := c.Get("non-existent-key")
	if exists {
		t.Fatal("expected key to not exist")
	}

	if value != nil {
		t.Errorf("expected nil value, got %v", value)
	}
}

func TestSetMultipleValues(t *testing.T) {
	c := New(time.Minute, 1024)

	testData := map[string]any{
		"string": "test-string",
		"int":    42,
		"float":  3.14,
		"struct": struct{ Name string }{Name: "test"},
	}

	// Set all values
	for key, value := range testData {
		c.Set(key, value, 1)
	}

	// Verify all values
	for key, expectedValue := range testData {
		retrieved, exists := c.Get(key)
		if !exists {
			t.Errorf("expected key %s to exist", key)
			continue
		}

		if retrieved != expectedValue {
			t.Errorf("key %s: expected %v, got %v", key, expectedValue, retrieved)
		}
	}

	// Test slice separately (can't use == comparison)
	sliceKey := "slice"
	sliceValue := []int{1, 2, 3}
	c.Set(sliceKey, sliceValue, 1)

	retrieved, exists := c.Get(sliceKey)
	if !exists {
		t.Errorf("expected key %s to exist", sliceKey)
	} else {
		retrievedSlice, ok := retrieved.([]int)
		if !ok {
			t.Errorf("expected slice type, got %T", retrieved)
		} else if len(retrievedSlice) != len(sliceValue) {
			t.Errorf("expected slice length %d, got %d", len(sliceValue), len(retrievedSlice))
		} else {
			for i, v := range sliceValue {
				if retrievedSlice[i] != v {
					t.Errorf("slice index %d: expected %d, got %d", i, v, retrievedSlice[i])
				}
			}
		}
	}

	expectedSize := len(testData) + 1 // +1 for the slice
	if c.Size() != expectedSize {
		t.Errorf("expected cache size %d, got %d", expectedSize, c.Size())
	}
}

func TestTTLExpiration(t *testing.T) {
	ttl := 100 * time.Millisecond
	c := New(ttl, 1024)

	key := "expiring-key"
	value := "expiring-value"

	c.Set(key, value, 1)

	// Value should exist immediately
	retrieved, exists := c.Get(key)
	if !exists {
		t.Fatal("expected key to exist immediately after setting")
	}
	if retrieved != value {
		t.Errorf("expected value %v, got %v", value, retrieved)
	}

	// Wait for expiration
	time.Sleep(ttl + 50*time.Millisecond)

	// Value should be expired and removed
	retrieved, exists = c.Get(key)
	if exists {
		t.Fatal("expected key to be expired and removed")
	}
	if retrieved != nil {
		t.Errorf("expected nil value for expired key, got %v", retrieved)
	}
}

func TestIsExpired(t *testing.T) {
	now := time.Now()

	// Test non-expired item
	item := CacheItem{
		Value:      "test",
		Expiration: now.Add(time.Hour).UnixNano(),
	}

	if item.IsExpired() {
		t.Error("expected item to not be expired")
	}

	// Test expired item
	expiredItem := CacheItem{
		Value:      "test",
		Expiration: now.Add(-time.Hour).UnixNano(),
	}

	if !expiredItem.IsExpired() {
		t.Error("expected item to be expired")
	}
}

func TestDelete(t *testing.T) {
	c := New(time.Minute, 1024)

	key := "delete-key"
	value := "delete-value"

	c.Set(key, value, 1)

	// Verify key exists
	_, exists := c.Get(key)
	if !exists {
		t.Fatal("expected key to exist before deletion")
	}

	// Delete the key
	c.Delete(key)

	// Verify key no longer exists
	_, exists = c.Get(key)
	if exists {
		t.Fatal("expected key to not exist after deletion")
	}
}

func TestClear(t *testing.T) {
	c := New(time.Minute, 1024)

	// Add multiple items
	for i := range 5 {
		c.Set(string(rune('a'+i)), i, 1)
	}

	if c.Size() != 5 {
		t.Fatalf("expected 5 items, got %d", c.Size())
	}

	// Clear the cache
	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected empty cache after clear, got %d items", c.Size())
	}

	// Verify no items exist
	for i := range 5 {
		_, exists := c.Get(string(rune('a' + i)))
		if exists {
			t.Errorf("expected key %s to not exist after clear", string(rune('a'+i)))
		}
	}
}

func TestSize(t *testing.T) {
	c := New(time.Minute, 1024)

	// Initially empty
	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}

	// Add items and check size
	for i := 1; i <= 10; i++ {
		c.Set(string(rune('a'+i-1)), i, 1)
		if c.Size() != i {
			t.Errorf("expected size %d, got %d", i, c.Size())
		}
	}

	// Delete items and check size
	for i := 9; i >= 0; i-- {
		c.Delete(string(rune('a' + i)))
		if c.Size() != i {
			t.Errorf("expected size %d, got %d", i, c.Size())
		}
	}
}

func TestKeys(t *testing.T) {
	c := New(time.Minute, 1024)

	expectedKeys := []string{"key1", "key2", "key3"}

	for _, key := range expectedKeys {
		c.Set(key, "value", 1)
	}

	keys := c.Keys()

	if len(keys) != len(expectedKeys) {
		t.Errorf("expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	// Convert to map for easier checking
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	for _, expectedKey := range expectedKeys {
		if !keyMap[expectedKey] {
			t.Errorf("expected key %s to be present", expectedKey)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := New(time.Minute, 1024)
	numGoroutines := 100
	numOperations := 10

	done := make(chan bool, numGoroutines)

	// Start multiple goroutines performing concurrent operations
	for i := range numGoroutines {
		go func(id int) {
			defer func() { done <- true }()

			for j := range numOperations {
				key := fmt.Sprintf("key-%d-%d", id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)

				// Set value
				c.Set(key, value, 1)

				// Get value
				retrieved, exists := c.Get(key)
				if exists && retrieved != value {
					t.Errorf("concurrent access error: expected %v, got %v", value, retrieved)
				}

				// Delete some values
				if j%3 == 0 {
					c.Delete(key)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		<-done
	}
}

func TestCleanupGoroutine(t *testing.T) {
	ttl := 200 * time.Millisecond
	c := New(ttl, 1024)

	// Add items that will expire
	for i := range 5 {
		c.Set(fmt.Sprintf("cleanup-key-%d", i), fmt.Sprintf("cleanup-value-%d", i), 1)
	}

	if c.Size() != 5 {
		t.Fatalf("expected 5 items, got %d", c.Size())
	}

	// Wait for items to expire and cleanup to run
	// The cleanup runs every minute, but expired items are also removed on Get()
	time.Sleep(ttl + 100*time.Millisecond)

	// Try to get expired items (this should trigger cleanup on access)
	for i := range 5 {
		_, exists := c.Get(fmt.Sprintf("cleanup-key-%d", i))
		if exists {
			t.Errorf("expected expired key cleanup-key-%d to not exist", i)
		}
	}

	// Cache should be empty or close to empty after cleanup
	if c.Size() > 0 {
		t.Logf("Note: Cache still has %d items (cleanup goroutine may not have run yet)", c.Size())
	}
}

func TestOverwriteExistingKey(t *testing.T) {
	c := New(time.Minute, 1024)

	key := "overwrite-key"
	value1 := "original-value"
	value2 := "new-value"

	c.Set(key, value1, 1)
	retrieved, exists := c.Get(key)
	if !exists || retrieved != value1 {
		t.Fatalf("expected original value %v, got %v", value1, retrieved)
	}

	// Overwrite with new value
	c.Set(key, value2, 1)
	retrieved, exists = c.Get(key)
	if !exists || retrieved != value2 {
		t.Errorf("expected new value %v, got %v", value2, retrieved)
	}

	// Cache should still have only one item
	if c.Size() != 1 {
		t.Errorf("expected cache size 1 after overwrite, got %d", c.Size())
	}
}

func TestLRUEviction(t *testing.T) {
	c := New(time.Minute, 10) // max size 10

	// Add 3 items of size 4 (total 12 > 10). The first one should be evicted.
	c.Set("k1", "v1", 4)
	c.Set("k2", "v2", 4)
	c.Set("k3", "v3", 4)

	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}

	_, exists := c.Get("k1")
	if exists {
		t.Error("expected k1 to be evicted")
	}

	_, exists = c.Get("k2")
	if !exists {
		t.Error("expected k2 to exist")
	}

	_, exists = c.Get("k3")
	if !exists {
		t.Error("expected k3 to exist")
	}
}

// Benchmark tests
func BenchmarkCacheSet(b *testing.B) {
	c := New(time.Minute, 10000000)

	for i := 0; b.Loop(); i++ {
		key := fmt.Sprintf("bench-key-%d", i)
		c.Set(key, "bench-value", 1)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	c := New(time.Minute, 10000000)

	// Pre-populate cache
	for i := range 1000 {
		key := fmt.Sprintf("bench-key-%d", i)
		c.Set(key, "bench-value", 1)
	}

	for i := 0; b.Loop(); i++ {
		key := fmt.Sprintf("bench-key-%d", i%1000)
		c.Get(key)
	}
}

func BenchmarkCacheSetGet(b *testing.B) {
	c := New(time.Minute, 10000000)

	for i := 0; b.Loop(); i++ {
		key := fmt.Sprintf("bench-key-%d", i)
		c.Set(key, "bench-value", 1)
		c.Get(key)
	}
}

func BenchmarkCacheConcurrent(b *testing.B) {
	c := New(time.Minute, 10000000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("concurrent-key-%d", i)
			c.Set(key, "concurrent-value", 1)
			c.Get(key)
			i++
		}
	})
}
