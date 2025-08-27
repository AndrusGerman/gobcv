// Package command contiene los comandos de la aplicación siguiendo el patrón CQRS.
package command

import (
	"context"
	"fmt"
	"log"

	"gobcv/internal/domain/repository"
	"gobcv/internal/domain/service"
)

// RefreshCurrenciesCommand representa el comando para actualizar monedas.
type RefreshCurrenciesCommand struct {
	ForceRefresh bool `json:"force_refresh"`
}

// RefreshCurrenciesHandler maneja el comando de actualización de monedas.
type RefreshCurrenciesHandler struct {
	currencyRepo repository.CurrencyRepository
	scraper      service.CurrencyScraper
	cache        service.CacheService
}

// NewRefreshCurrenciesHandler crea un nuevo handler para el comando.
func NewRefreshCurrenciesHandler(
	currencyRepo repository.CurrencyRepository,
	scraper service.CurrencyScraper,
	cache service.CacheService,
) *RefreshCurrenciesHandler {
	return &RefreshCurrenciesHandler{
		currencyRepo: currencyRepo,
		scraper:      scraper,
		cache:        cache,
	}
}

// RefreshCurrenciesResult representa el resultado del comando.
type RefreshCurrenciesResult struct {
	UpdatedCount int      `json:"updated_count"`
	Currencies   []string `json:"currencies"`
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
}

// Handle ejecuta el comando de actualización de monedas.
func (h *RefreshCurrenciesHandler) Handle(ctx context.Context, cmd RefreshCurrenciesCommand) (*RefreshCurrenciesResult, error) {
	log.Printf("Ejecutando comando RefreshCurrencies (force_refresh: %v)", cmd.ForceRefresh)

	// Verificar si el scraper está disponible
	if err := h.scraper.IsHealthy(ctx); err != nil {
		return &RefreshCurrenciesResult{
			Success: false,
			Message: fmt.Sprintf("Scraper no disponible: %v", err),
		}, err
	}

	// Obtener monedas desde la fuente externa
	currencies, err := h.scraper.ScrapeCurrencies(ctx)
	if err != nil {
		return &RefreshCurrenciesResult{
			Success: false,
			Message: fmt.Sprintf("Error al obtener monedas: %v", err),
		}, err
	}

	var updatedCurrencies []string

	// Guardar cada moneda en el repositorio
	for _, currency := range currencies {
		if err := h.currencyRepo.Save(ctx, currency); err != nil {
			log.Printf("Error al guardar moneda %s: %v", currency.ID, err)
			continue
		}

		// Invalidar caché para esta moneda
		cacheKey := fmt.Sprintf("currency:%s", currency.ID)
		h.cache.Delete(ctx, cacheKey)

		updatedCurrencies = append(updatedCurrencies, currency.ID)
	}

	// Invalidar caché de listado general
	h.cache.Delete(ctx, "currencies:all")

	return &RefreshCurrenciesResult{
		UpdatedCount: len(updatedCurrencies),
		Currencies:   updatedCurrencies,
		Success:      true,
		Message:      fmt.Sprintf("Se actualizaron %d monedas exitosamente", len(updatedCurrencies)),
	}, nil
}
