// Package cache implementa el adaptador de caché en memoria.
package cache

import (
	"context"
	"sync"
	"time"

	"gobcv/internal/domain/service"
)

// cacheItem representa un elemento en el caché con su tiempo de expiración.
type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// isExpired verifica si el elemento ha expirado.
func (item *cacheItem) isExpired() bool {
	return time.Now().After(item.expiresAt)
}

// MemoryCache implementa el servicio de caché en memoria.
type MemoryCache struct {
	items   map[string]*cacheItem
	mutex   sync.RWMutex
	stats   service.CacheStats
	cleaner *time.Ticker
	stopCh  chan bool
}

// NewMemoryCache crea una nueva instancia del caché en memoria.
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items:  make(map[string]*cacheItem),
		stats:  service.CacheStats{},
		stopCh: make(chan bool),
	}

	// Iniciar limpieza automática cada 5 minutos
	cache.startCleanup()

	return cache
}

// startCleanup inicia la rutina de limpieza automática de elementos expirados.
func (c *MemoryCache) startCleanup() {
	c.cleaner = time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-c.cleaner.C:
				c.cleanup()
			case <-c.stopCh:
				c.cleaner.Stop()
				return
			}
		}
	}()
}

// cleanup elimina los elementos expirados del caché.
func (c *MemoryCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, item := range c.items {
		if item.isExpired() {
			delete(c.items, key)
		}
	}
}

// Set almacena un valor en el caché con una clave y TTL especificados.
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiresAt := time.Now().Add(ttl)
	c.items[key] = &cacheItem{
		value:     value,
		expiresAt: expiresAt,
	}

	return nil
}

// Get obtiene un valor del caché por su clave.
func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, nil
	}

	if item.isExpired() {
		c.stats.Misses++
		// Eliminar el elemento expirado
		go func() {
			c.mutex.Lock()
			delete(c.items, key)
			c.mutex.Unlock()
		}()
		return nil, nil
	}

	c.stats.Hits++
	return item.value, nil
}

// Delete elimina un valor del caché.
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
	return nil
}

// Exists verifica si una clave existe en el caché.
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false, nil
	}

	if item.isExpired() {
		// Eliminar el elemento expirado
		go func() {
			c.mutex.Lock()
			delete(c.items, key)
			c.mutex.Unlock()
		}()
		return false, nil
	}

	return true, nil
}

// Clear limpia todo el caché.
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*cacheItem)
	c.stats = service.CacheStats{}
	return nil
}

// GetStats obtiene estadísticas del caché.
func (c *MemoryCache) GetStats(ctx context.Context) (service.CacheStats, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := c.stats
	stats.Keys = int64(len(c.items))

	return stats, nil
}

// Close cierra el caché y detiene la rutina de limpieza.
func (c *MemoryCache) Close() {
	close(c.stopCh)
}
