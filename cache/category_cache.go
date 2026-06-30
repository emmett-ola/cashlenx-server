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
	byScope   map[string]*model.CategoryEntity
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
			byScope:   make(map[string]*model.CategoryEntity),
			enabled:   true,
			lastClear: time.Now(),
		}
		util.Logger.Info("Category cache initialized")
	})
	return instance
}

// GetByName retrieves a category by name from cache
func (c *CategoryCache) GetByName(name string) (*model.CategoryEntity, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.enabled {
		return nil, false
	}
	entity, ok := c.byName[name]
	if ok {
		c.hits++
		util.Logger.Debugw("Category cache hit", "name", name)
	} else {
		c.misses++
		util.Logger.Debugw("Category cache miss", "name", name)
	}
	return cloneCategory(entity), ok
}

// GetByID retrieves a category by ID from cache
func (c *CategoryCache) GetByID(id string) (*model.CategoryEntity, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.enabled {
		return nil, false
	}
	entity, ok := c.byID[id]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return cloneCategory(entity), ok
}

// GetByScope retrieves an unambiguous user/type/parent/name category lookup.
func (c *CategoryCache) GetByScope(userID, categoryType, parentID, name string) (*model.CategoryEntity, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.enabled {
		return nil, false
	}
	key := categoryScopeKey(userID, categoryType, parentID, name)
	entity, ok := c.byScope[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return cloneCategory(entity), ok
}

// Set adds or updates a category in the cache
func (c *CategoryCache) Set(entity *model.CategoryEntity) {
	if entity == nil || entity.IsEmpty() {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.enabled {
		return
	}

	copy := *entity
	c.byName[entity.Name] = &copy
	c.byID[entity.Id.Hex()] = &copy
	c.byScope[categoryScopeKey(entity.BelongsUserId.Hex(), entity.Type, entity.ParentId.Hex(), entity.Name)] = &copy
	util.Logger.Debugw("Category cached", "name", entity.Name, "id", entity.Id.Hex())
}

// Invalidate removes a category from cache by name
func (c *CategoryCache) Invalidate(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entity, ok := c.byName[name]; ok {
		delete(c.byID, entity.Id.Hex())
		delete(c.byScope, categoryScopeKey(entity.BelongsUserId.Hex(), entity.Type, entity.ParentId.Hex(), entity.Name))
		util.Logger.Debugw("Category invalidated", "name", name)
	}
	delete(c.byName, name)
}

// InvalidateByID removes a category from cache by ID
func (c *CategoryCache) InvalidateByID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entity, ok := c.byID[id]; ok {
		if named, exists := c.byName[entity.Name]; exists && named.Id.Hex() == id {
			delete(c.byName, entity.Name)
		}
		delete(c.byScope, categoryScopeKey(entity.BelongsUserId.Hex(), entity.Type, entity.ParentId.Hex(), entity.Name))
		util.Logger.Debugw("Category invalidated", "id", id)
	}
	delete(c.byID, id)
}

// InvalidateUser removes only entries owned by one user.
func (c *CategoryCache) InvalidateUser(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, entity := range c.byID {
		if entity.BelongsUserId.Hex() != userID {
			continue
		}
		delete(c.byID, id)
		delete(c.byScope, categoryScopeKey(entity.BelongsUserId.Hex(), entity.Type, entity.ParentId.Hex(), entity.Name))
		if named, exists := c.byName[entity.Name]; exists && named.Id == entity.Id {
			delete(c.byName, entity.Name)
		}
	}
}

// Clear removes all categories from cache
func (c *CategoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.byName = make(map[string]*model.CategoryEntity)
	c.byID = make(map[string]*model.CategoryEntity)
	c.byScope = make(map[string]*model.CategoryEntity)
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
		"size":       len(c.byID),
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
	c.byScope = make(map[string]*model.CategoryEntity)
	util.Logger.Info("Category cache disabled")
}

func categoryScopeKey(userID, categoryType, parentID, name string) string {
	return userID + "\x00" + categoryType + "\x00" + parentID + "\x00" + name
}

func cloneCategory(entity *model.CategoryEntity) *model.CategoryEntity {
	if entity == nil {
		return nil
	}
	copy := *entity
	return &copy
}
