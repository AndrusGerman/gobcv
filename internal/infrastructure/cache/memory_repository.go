// Package cache implementa el repositorio en memoria para monedas.
package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gobcv/internal/domain/entity"
	"gobcv/internal/domain/repository"
)

// MemoryRepository implementa el repositorio de monedas en memoria.
type MemoryRepository struct {
	currencies map[string]*entity.Currency
	mutex      sync.RWMutex
}

// NewMemoryRepository crea un nuevo repositorio en memoria.
func NewMemoryRepository() repository.CurrencyRepository {
	return &MemoryRepository{
		currencies: make(map[string]*entity.Currency),
	}
}

// Save guarda o actualiza una moneda en el repositorio.
func (r *MemoryRepository) Save(ctx context.Context, currency *entity.Currency) error {
	if currency == nil {
		return fmt.Errorf("currency cannot be nil")
	}

	if !currency.IsValid() {
		return fmt.Errorf("currency is not valid")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()

	existing, exists := r.currencies[currency.ID]
	if exists {
		// Preservar el historial existente
		currency.History = existing.History
		currency.PreviousValue = existing.PreviousValue
		currency.PreviousUpdatedAt = existing.PreviousUpdatedAt
		currency.Difference = existing.Difference

		// Si el valor es diferente al anterior, actualizar el historial
		if existing.Value != currency.Value {
			difference := currency.Value - existing.Value

			// Agregar el valor anterior al historial
			historyEntry := entity.CurrencyHistory{
				Value:      existing.Value,
				UpdatedAt:  existing.UpdatedAt,
				Difference: existing.Difference,
			}

			// Agregar al inicio del historial (limitado a 10 elementos)
			newHistory := append([]entity.CurrencyHistory{historyEntry}, currency.History...)
			if len(newHistory) > 10 {
				newHistory = newHistory[:10]
			}

			currency.History = newHistory
			currency.PreviousValue = existing.Value
			currency.PreviousUpdatedAt = existing.UpdatedAt
			currency.Difference = difference
		}
	} else {
		// Primera vez, asegurar que History esté inicializado
		if currency.History == nil {
			currency.History = make([]entity.CurrencyHistory, 0)
		}
	}

	// Actualizar timestamp
	currency.UpdatedAt = now

	// Crear una copia para evitar modificaciones externas
	currencyCopy := *currency
	r.currencies[currency.ID] = &currencyCopy

	return nil
}

// FindByID busca una moneda por su ID.
func (r *MemoryRepository) FindByID(ctx context.Context, id string) (*entity.Currency, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	currency, exists := r.currencies[id]
	if !exists {
		return nil, nil
	}

	// Retornar una copia para evitar modificaciones externas
	currencyCopy := *currency
	return &currencyCopy, nil
}

// FindAll obtiene todas las monedas disponibles.
func (r *MemoryRepository) FindAll(ctx context.Context) ([]*entity.Currency, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	currencies := make([]*entity.Currency, 0, len(r.currencies))

	for _, currency := range r.currencies {
		// Crear copia para evitar modificaciones externas
		currencyCopy := *currency
		currencies = append(currencies, &currencyCopy)
	}

	return currencies, nil
}

// Delete elimina una moneda del repositorio.
func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.currencies, id)
	return nil
}

// FindByLastUpdate busca monedas actualizadas después de una fecha específica.
func (r *MemoryRepository) FindByLastUpdate(ctx context.Context, since time.Time) ([]*entity.Currency, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var currencies []*entity.Currency

	for _, currency := range r.currencies {
		if currency.UpdatedAt.After(since) {
			// Crear copia para evitar modificaciones externas
			currencyCopy := *currency
			currencies = append(currencies, &currencyCopy)
		}
	}

	return currencies, nil
}

// Count retorna la cantidad de monedas almacenadas.
func (r *MemoryRepository) Count(ctx context.Context) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.currencies), nil
}
