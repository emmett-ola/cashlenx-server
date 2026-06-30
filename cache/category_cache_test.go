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

	// Get by ID
	retrieved, ok := cache.GetByID(entity.Id.Hex())
	if !ok {
		t.Error("Expected to find category by ID")
	}
	if retrieved.Id != entity.Id {
		t.Error("Expected same ID")
	}
}

func TestCategoryCache_InvalidateByID(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()

	entity := &model.CategoryEntity{
		Id:   primitive.NewObjectID(),
		Name: "TestCategory",
	}

	cache.Set(entity)

	// Verify it's cached
	_, ok := cache.GetByID(entity.Id.Hex())
	if !ok {
		t.Error("Expected category to be cached")
	}

	// Invalidate
	cache.InvalidateByID(entity.Id.Hex())

	// Verify it's removed
	_, ok = cache.GetByID(entity.Id.Hex())
	if ok {
		t.Error("Expected category to be removed from cache")
	}
}

func TestCategoryCache_UserScopedLookup(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()
	first := &model.CategoryEntity{Id: primitive.NewObjectID(), BelongsUserId: primitive.NewObjectID(), Name: "Food", Type: "expense"}
	second := &model.CategoryEntity{Id: primitive.NewObjectID(), BelongsUserId: primitive.NewObjectID(), Name: "Food", Type: "expense"}
	cache.Set(first)
	cache.Set(second)
	got, ok := cache.GetByScope(first.BelongsUserId.Hex(), first.Type, first.ParentId.Hex(), first.Name)
	if !ok || got.Id != first.Id {
		t.Fatalf("scoped lookup returned %#v", got)
	}
	got.Name = "mutated"
	again, _ := cache.GetByID(first.Id.Hex())
	if again.Name != "Food" {
		t.Fatal("cache returned mutable internal entity")
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
	cache.GetByID(entity.Id.Hex())               // hit
	cache.GetByID(entity.Id.Hex())               // hit
	cache.GetByID(primitive.NewObjectID().Hex()) // miss

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
	_, ok := cache.GetByID(entity.Id.Hex())
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

	retrieved, ok := cache2.GetByID(entity.Id.Hex())
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
					_, _ = cache.GetByID(entities[idx].Id.Hex())
				case 1: // Write
					idx := j % len(entities)
					cache.Set(entities[idx])
				case 2: // Invalidate
					idx := j % len(entities)
					cache.InvalidateByID(entities[idx].Id.Hex())
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
	_, ok := cache.GetByID(entity.Id.Hex())
	if !ok {
		t.Error("Expected category to be cached")
	}

	// Invalidate from one goroutine
	done := make(chan bool)
	go func() {
		cache.InvalidateByID(entity.Id.Hex())
		done <- true
	}()
	<-done

	// Verify invalidation is visible from main goroutine
	_, ok = cache.GetByID(entity.Id.Hex())
	if ok {
		t.Error("Expected category to be invalidated across goroutines")
	}
}

func TestCategoryCache_ClearAcrossThreads(t *testing.T) {
	cache := GetCategoryCache()
	cache.Clear()

	// Add multiple categories
	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		entity := &model.CategoryEntity{
			Id:   primitive.NewObjectID(),
			Name: "Category" + string(rune('A'+i)),
		}
		cache.Set(entity)
		ids = append(ids, entity.Id.Hex())
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
	for _, id := range ids {
		_, ok := cache.GetByID(id)
		if ok {
			t.Errorf("Expected category %s to be cleared", id)
		}
	}
}
