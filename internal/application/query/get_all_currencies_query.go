// Package query contiene las consultas para obtener múltiples monedas.
package query

import (
	"context"
	"encoding/json"
	"time"

	"gobcv/internal/domain/entity"
	"gobcv/internal/domain/repository"
	"gobcv/internal/domain/service"
)

// GetAllCurrenciesQuery representa la consulta para obtener todas las monedas.
type GetAllCurrenciesQuery struct {
	UseCache     bool `json:"use_cache"`
	IncludeStale bool `json:"include_stale"`
}

// GetAllCurrenciesHandler maneja las consultas de múltiples monedas.
type GetAllCurrenciesHandler struct {
	currencyRepo repository.CurrencyRepository
	cache        service.CacheService
}

// NewGetAllCurrenciesHandler crea un nuevo handler para consultas de múltiples monedas.
func NewGetAllCurrenciesHandler(
	currencyRepo repository.CurrencyRepository,
	cache service.CacheService,
) *GetAllCurrenciesHandler {
	return &GetAllCurrenciesHandler{
		currencyRepo: currencyRepo,
		cache:        cache,
	}
}

// GetAllCurrenciesResult representa el resultado de la consulta.
type GetAllCurrenciesResult struct {
	Currencies []*entity.Currency `json:"currencies"`
	Count      int                `json:"count"`
	FromCache  bool               `json:"from_cache"`
	Success    bool               `json:"success"`
	Message    string             `json:"message"`
}

// Handle ejecuta la consulta para obtener todas las monedas.
func (h *GetAllCurrenciesHandler) Handle(ctx context.Context, query GetAllCurrenciesQuery) (*GetAllCurrenciesResult, error) {
	cacheKey := "currencies:all"

	// Intentar obtener desde caché si está habilitado
	if query.UseCache {
		if cached, err := h.cache.Get(ctx, cacheKey); err == nil && cached != nil {
			if currenciesData, ok := cached.([]byte); ok {
				var currencies []*entity.Currency
				if err := json.Unmarshal(currenciesData, &currencies); err == nil {
					return &GetAllCurrenciesResult{
						Currencies: currencies,
						Count:      len(currencies),
						FromCache:  true,
						Success:    true,
						Message:    "Monedas obtenidas desde caché",
					}, nil
				}
			}
		}
	}

	// Obtener desde repositorio
	currencies, err := h.currencyRepo.FindAll(ctx)
	if err != nil {
		return &GetAllCurrenciesResult{
			Success: false,
			Message: "Error al obtener monedas desde repositorio",
		}, err
	}

	// Filtrar monedas obsoletas si no se incluyen
	if !query.IncludeStale {
		var freshCurrencies []*entity.Currency
		for _, currency := range currencies {
			if !currency.IsStale(30 * time.Minute) { // Considera obsoleto después de 30 minutos
				freshCurrencies = append(freshCurrencies, currency)
			}
		}
		currencies = freshCurrencies
	}

	// Guardar en caché si está habilitado
	if query.UseCache {
		if currenciesData, err := json.Marshal(currencies); err == nil {
			// Cache por 2 minutos
			h.cache.Set(ctx, cacheKey, currenciesData, 2*time.Minute)
		}
	}

	return &GetAllCurrenciesResult{
		Currencies: currencies,
		Count:      len(currencies),
		FromCache:  false,
		Success:    true,
		Message:    "Monedas obtenidas desde repositorio",
	}, nil
}
