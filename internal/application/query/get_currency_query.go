// Package query contiene las consultas de la aplicación siguiendo el patrón CQRS.
package query

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gobcv/internal/domain/entity"
	"gobcv/internal/domain/repository"
	"gobcv/internal/domain/service"
)

// GetCurrencyQuery representa la consulta para obtener una moneda específica.
type GetCurrencyQuery struct {
	CurrencyID string `json:"currency_id"`
	UseCache   bool   `json:"use_cache"`
}

// GetCurrencyHandler maneja las consultas de monedas individuales.
type GetCurrencyHandler struct {
	currencyRepo repository.CurrencyRepository
	cache        service.CacheService
}

// NewGetCurrencyHandler crea un nuevo handler para consultas de monedas.
func NewGetCurrencyHandler(
	currencyRepo repository.CurrencyRepository,
	cache service.CacheService,
) *GetCurrencyHandler {
	return &GetCurrencyHandler{
		currencyRepo: currencyRepo,
		cache:        cache,
	}
}

// GetCurrencyResult representa el resultado de la consulta.
type GetCurrencyResult struct {
	Currency   *entity.Currency `json:"currency"`
	FromCache  bool             `json:"from_cache"`
	Success    bool             `json:"success"`
	Message    string           `json:"message"`
}

// Handle ejecuta la consulta para obtener una moneda.
func (h *GetCurrencyHandler) Handle(ctx context.Context, query GetCurrencyQuery) (*GetCurrencyResult, error) {
	cacheKey := fmt.Sprintf("currency:%s", query.CurrencyID)
	
	// Intentar obtener desde caché si está habilitado
	if query.UseCache {
		if cached, err := h.cache.Get(ctx, cacheKey); err == nil && cached != nil {
			if currencyData, ok := cached.([]byte); ok {
				var currency entity.Currency
				if err := json.Unmarshal(currencyData, &currency); err == nil {
					return &GetCurrencyResult{
						Currency:  &currency,
						FromCache: true,
						Success:   true,
						Message:   "Moneda obtenida desde caché",
					}, nil
				}
			}
		}
	}

	// Obtener desde repositorio
	currency, err := h.currencyRepo.FindByID(ctx, query.CurrencyID)
	if err != nil {
		return &GetCurrencyResult{
			Success: false,
			Message: fmt.Sprintf("Error al obtener moneda: %v", err),
		}, err
	}

	if currency == nil {
		return &GetCurrencyResult{
			Success: false,
			Message: "Moneda no encontrada",
		}, nil
	}

	// Guardar en caché si está habilitado
	if query.UseCache {
		if currencyData, err := json.Marshal(currency); err == nil {
			// Cache por 5 minutos
			h.cache.Set(ctx, cacheKey, currencyData, 5*time.Minute)
		}
	}

	return &GetCurrencyResult{
		Currency:  currency,
		FromCache: false,
		Success:   true,
		Message:   "Moneda obtenida desde repositorio",
	}, nil
}
