package cache

import (
	"sync"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

// CategoryCache provides thread-safe in-memory caching for categories
type CategoryCache struct {
	byName    map[string]*model.CategoryEntity
	byID      map[string]*model.CategoryEntity
	mu        sync.RWMutex
	hits      int64
	misses    int64
	enabled   bool
	lastClear time.Time
}

var (
	instance *CategoryCache
	once     sync.Once
)

// GetCategoryCache returns the singleton category cache instance
func GetCategoryCache() *CategoryCache {
	once.Do(func() {
		instance = &CategoryCache{
			byName:    make(map[string]*model.CategoryEntity),
			byID:      make(map[string]*model.CategoryEntity),
			enabled:   true,
			lastClear: time.Now(),
		}
		util.Logger.Info("Category cache initialized")
	})
	return instance
}

// GetByName retrieves a category by name from cache
func (c *CategoryCache) GetByName(name string) (*model.CategoryEntity, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	entity, ok := c.byName[name]
	c.mu.RUnlock()

	// Update stats outside of read lock to avoid race
	c.mu.Lock()
	if ok {
		c.hits++
		util.Logger.Debugw("Category cache hit", "name", name)
	} else {
		c.misses++
		util.Logger.Debugw("Category cache miss", "name", name)
	}
	c.mu.Unlock()

	return entity, ok
}

// GetByID retrieves a category by ID from cache
func (c *CategoryCache) GetByID(id string) (*model.CategoryEntity, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	entity, ok := c.byID[id]
	c.mu.RUnlock()

	// Update stats outside of read lock to avoid race
	c.mu.Lock()
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	c.mu.Unlock()

	return entity, ok
}

// Set adds or updates a category in the cache
func (c *CategoryCache) Set(entity *model.CategoryEntity) {
	if !c.enabled || entity == nil || entity.IsEmpty() {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.byName[entity.Name] = entity
	c.byID[entity.Id.Hex()] = entity
	util.Logger.Debugw("Category cached", "name", entity.Name, "id", entity.Id.Hex())
}

// Invalidate removes a category from cache by name
func (c *CategoryCache) Invalidate(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entity, ok := c.byName[name]; ok {
		delete(c.byID, entity.Id.Hex())
		util.Logger.Debugw("Category invalidated", "name", name)
	}
	delete(c.byName, name)
}

// InvalidateByID removes a category from cache by ID
func (c *CategoryCache) InvalidateByID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entity, ok := c.byID[id]; ok {
		delete(c.byName, entity.Name)
		util.Logger.Debugw("Category invalidated", "id", id)
	}
	delete(c.byID, id)
}

// Clear removes all categories from cache
func (c *CategoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.byName = make(map[string]*model.CategoryEntity)
	c.byID = make(map[string]*model.CategoryEntity)
	c.lastClear = time.Now()
	util.Logger.Info("Category cache cleared")
}

// GetStats returns cache statistics
func (c *CategoryCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"enabled":    c.enabled,
		"size":       len(c.byName),
		"hits":       c.hits,
		"misses":     c.misses,
		"hit_rate":   hitRate,
		"last_clear": c.lastClear,
	}
}

// Enable enables the cache
func (c *CategoryCache) Enable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = true
	util.Logger.Info("Category cache enabled")
}

// Disable disables the cache and clears it
func (c *CategoryCache) Disable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = false
	c.byName = make(map[string]*model.CategoryEntity)
	c.byID = make(map[string]*model.CategoryEntity)
	util.Logger.Info("Category cache disabled")
}
