package cache

import (
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCategoryCache_SetAndGet(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()

	entity := &model.CategoryEntity{
		Id:   primitive.NewObjectID(),
		Name: "TestCategory",
	}

	// Set category
	cache.Set(entity)

	// Get by name
	retrieved, ok := cache.GetByName("TestCategory")
	if !ok {
		t.Error("Expected to find category by name")
	}
	if retrieved.Name != "TestCategory" {
		t.Errorf("Expected name 'TestCategory', got '%s'", retrieved.Name)
	}

	// Get by ID
	retrieved, ok = cache.GetByID(entity.Id.Hex())
	if !ok {
		t.Error("Expected to find category by ID")
	}
	if retrieved.Id != entity.Id {
		t.Error("Expected same ID")
	}
}

func TestCategoryCache_Invalidate(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()

	entity := &model.CategoryEntity{
		Id:   primitive.NewObjectID(),
		Name: "TestCategory",
	}

	cache.Set(entity)

	// Verify it's cached
	_, ok := cache.GetByName("TestCategory")
	if !ok {
		t.Error("Expected category to be cached")
	}

	// Invalidate
	cache.Invalidate("TestCategory")

	// Verify it's removed
	_, ok = cache.GetByName("TestCategory")
	if ok {
		t.Error("Expected category to be removed from cache")
	}
}

func TestCategoryCache_Clear(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()

	// Add multiple categories
	for i := 0; i < 5; i++ {
		entity := &model.CategoryEntity{
			Id:   primitive.NewObjectID(),
			Name: "Category" + string(rune(i)),
		}
		cache.Set(entity)
	}

	stats := cache.GetStats()
	if stats["size"].(int) != 5 {
		t.Errorf("Expected cache size 5, got %d", stats["size"])
	}

	// Clear cache
	cache.Clear()

	stats = cache.GetStats()
	if stats["size"].(int) != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", stats["size"])
	}
}

func TestCategoryCache_Stats(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()

	entity := &model.CategoryEntity{
		Id:   primitive.NewObjectID(),
		Name: "TestCategory",
	}
	cache.Set(entity)

	// Generate some hits and misses
	cache.GetByName("TestCategory") // hit
	cache.GetByName("TestCategory") // hit
	cache.GetByName("NonExistent")  // miss

	stats := cache.GetStats()

	if stats["hits"].(int64) < 2 {
		t.Errorf("Expected at least 2 hits, got %d", stats["hits"])
	}

	if stats["misses"].(int64) < 1 {
		t.Errorf("Expected at least 1 miss, got %d", stats["misses"])
	}

	hitRate := stats["hit_rate"].(float64)
	if hitRate <= 0 || hitRate > 100 {
		t.Errorf("Expected hit rate between 0 and 100, got %f", hitRate)
	}
}

func TestCategoryCache_Disable(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()
	cache.Enable()

	entity := &model.CategoryEntity{
		Id:   primitive.NewObjectID(),
		Name: "TestCategory",
	}
	cache.Set(entity)

	// Disable cache
	cache.Disable()

	// Try to get - should return false
	_, ok := cache.GetByName("TestCategory")
	if ok {
		t.Error("Expected cache to be disabled")
	}

	// Re-enable for other tests
	cache.Enable()
}

func TestCategoryCache_Singleton(t *testing.T) {
	// Get cache instance multiple times
	cache1 := GetCategoryCache()
	cache2 := GetCategoryCache()

	// Should be the same instance
	if cache1 != cache2 {
		t.Error("Expected GetCategoryCache to return singleton instance")
	}

	// Set in one, should be visible in other
	entity := &model.CategoryEntity{
		Id:   primitive.NewObjectID(),
		Name: "SingletonTest",
	}
	cache1.Set(entity)

	retrieved, ok := cache2.GetByName("SingletonTest")
	if !ok {
		t.Error("Expected to find category set via cache1 in cache2")
	}
	if retrieved.Name != "SingletonTest" {
		t.Errorf("Expected name 'SingletonTest', got '%s'", retrieved.Name)
	}
}

func TestCategoryCache_ConcurrentAccess(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()

	// Number of concurrent goroutines
	numGoroutines := 100
	numOperations := 100

	// Create test entities
	entities := make([]*model.CategoryEntity, 10)
	for i := 0; i < 10; i++ {
		entities[i] = &model.CategoryEntity{
			Id:   primitive.NewObjectID(),
			Name: "Category" + string(rune('A'+i)),
		}
	}

	// Pre-populate cache
	for _, entity := range entities {
		cache.Set(entity)
	}

	// Channel to collect errors
	errChan := make(chan error, numGoroutines)
	doneChan := make(chan bool, numGoroutines)

	// Launch concurrent readers and writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { doneChan <- true }()

			for j := 0; j < numOperations; j++ {
				// Mix of operations
				switch j % 4 {
				case 0: // Read
					idx := j % len(entities)
					_, _ = cache.GetByName(entities[idx].Name)
				case 1: // Write
					idx := j % len(entities)
					cache.Set(entities[idx])
				case 2: // Invalidate
					idx := j % len(entities)
					cache.Invalidate(entities[idx].Name)
				case 3: // Get stats
					_ = cache.GetStats()
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-doneChan
	}
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Errorf("Concurrent access error: %v", err)
	}

	// Verify cache is still functional
	stats := cache.GetStats()
	if stats["size"].(int) < 0 {
		t.Error("Cache size should not be negative after concurrent access")
	}
}

func TestCategoryCache_InvalidationAcrossThreads(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()

	entity := &model.CategoryEntity{
		Id:   primitive.NewObjectID(),
		Name: "SharedCategory",
	}
	cache.Set(entity)

	// Verify it's cached
	_, ok := cache.GetByName("SharedCategory")
	if !ok {
		t.Error("Expected category to be cached")
	}

	// Invalidate from one goroutine
	done := make(chan bool)
	go func() {
		cache.Invalidate("SharedCategory")
		done <- true
	}()
	<-done

	// Verify invalidation is visible from main goroutine
	_, ok = cache.GetByName("SharedCategory")
	if ok {
		t.Error("Expected category to be invalidated across goroutines")
	}
}

func TestCategoryCache_ClearAcrossThreads(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()

	// Add multiple categories
	for i := 0; i < 5; i++ {
		entity := &model.CategoryEntity{
			Id:   primitive.NewObjectID(),
			Name: "Category" + string(rune('A'+i)),
		}
		cache.Set(entity)
	}

	stats := cache.GetStats()
	if stats["size"].(int) != 5 {
		t.Errorf("Expected cache size 5, got %d", stats["size"])
	}

	// Clear from another goroutine
	done := make(chan bool)
	go func() {
		cache.Clear()
		done <- true
	}()
	<-done

	// Verify clear is visible from main goroutine
	stats = cache.GetStats()
	if stats["size"].(int) != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", stats["size"])
	}

	// Verify all categories are gone
	for i := 0; i < 5; i++ {
		name := "Category" + string(rune('A'+i))
		_, ok := cache.GetByName(name)
		if ok {
			t.Errorf("Expected category %s to be cleared", name)
		}
	}
}
