// Package service define los puertos para servicios de cache.
package service

import (
	"context"
	"time"
)

// CacheService define el puerto para el servicio de caché.
type CacheService interface {
	// Set almacena un valor en el caché con una clave y TTL especificados.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// Get obtiene un valor del caché por su clave.
	Get(ctx context.Context, key string) (interface{}, error)
	
	// Delete elimina un valor del caché.
	Delete(ctx context.Context, key string) error
	
	// Exists verifica si una clave existe en el caché.
	Exists(ctx context.Context, key string) (bool, error)
	
	// Clear limpia todo el caché.
	Clear(ctx context.Context) error
	
	// GetStats obtiene estadísticas del caché.
	GetStats(ctx context.Context) (CacheStats, error)
}

// CacheStats representa las estadísticas del caché.
type CacheStats struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
	Keys   int64 `json:"keys"`
}
