package cache

import (
	"log"
	"sync"
	"time"

	"scalingo-api-test/internal/models"
)

// Cache represents an in-memory cache for storing CustomRepository objects
type Cache struct {
	repos     []*models.CustomRepository
	lastFetch time.Time
	mu        sync.RWMutex
}

// NewCache creates and returns a new instance of Cache
func NewCache() *Cache {
	return &Cache{}
}

// Set updates the cache with a new slice of CustomRepository objects
// and updates the lastFetch timestamp
func (c *Cache) Set(repos []*models.CustomRepository) {
	c.mu.Lock()         // Acquire write lock
	defer c.mu.Unlock() // Ensure lock is released after function execution

	c.repos = repos
	c.lastFetch = time.Now()

	log.Printf("Cache set with %d repositories", len(repos))
}

// Get retrieves the cached CustomRepository objects and the last fetch timestamp
func (c *Cache) Get() ([]*models.CustomRepository, time.Time) {
	c.mu.RLock()         // Acquire read lock
	defer c.mu.RUnlock() // Ensure lock is released after function execution

	return c.repos, c.lastFetch // Return cached data and timestamp
}
