// Package service contiene los servicios de aplicación que coordinan la lógica de negocio.
package service

import (
	"context"
	"log"
	"time"

	"gobcv/internal/application/command"
	"gobcv/internal/application/query"
	"gobcv/internal/domain/repository"
	"gobcv/internal/domain/service"
)

// CurrencyService coordina las operaciones relacionadas con monedas.
type CurrencyService struct {
	refreshHandler     *command.RefreshCurrenciesHandler
	getCurrencyHandler *query.GetCurrencyHandler
	getAllHandler      *query.GetAllCurrenciesHandler
	cacheService       service.CacheService
}

// NewCurrencyService crea una nueva instancia del servicio de monedas.
func NewCurrencyService(
	currencyRepo repository.CurrencyRepository,
	scraper service.CurrencyScraper,
	cache service.CacheService,
) *CurrencyService {
	return &CurrencyService{
		refreshHandler:     command.NewRefreshCurrenciesHandler(currencyRepo, scraper, cache),
		getCurrencyHandler: query.NewGetCurrencyHandler(currencyRepo, cache),
		getAllHandler:      query.NewGetAllCurrenciesHandler(currencyRepo, cache),
		cacheService:       cache,
	}
}

// GetRefreshHandler retorna el handler de refresh.
func (s *CurrencyService) GetRefreshHandler() *command.RefreshCurrenciesHandler {
	return s.refreshHandler
}

// GetCurrencyHandler retorna el handler de consulta de moneda individual.
func (s *CurrencyService) GetCurrencyHandler() *query.GetCurrencyHandler {
	return s.getCurrencyHandler
}

// GetAllCurrenciesHandler retorna el handler de consulta de todas las monedas.
func (s *CurrencyService) GetAllCurrenciesHandler() *query.GetAllCurrenciesHandler {
	return s.getAllHandler
}

// StartPeriodicRefresh inicia la actualización periódica de monedas.
func (s *CurrencyService) StartPeriodicRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Iniciando actualización periódica de monedas cada %v", interval)

	// Actualización inicial
	s.refreshCurrencies(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Deteniendo actualización periódica de monedas")
			return
		case <-ticker.C:
			s.refreshCurrencies(ctx)
		}
	}
}

// refreshCurrencies ejecuta la actualización de monedas y maneja errores.
func (s *CurrencyService) refreshCurrencies(ctx context.Context) {
	cmd := command.RefreshCurrenciesCommand{
		ForceRefresh: false,
	}

	result, err := s.refreshHandler.Handle(ctx, cmd)
	if err != nil {
		log.Printf("Error en actualización automática de monedas: %v", err)
		return
	}

	if result.Success {
		log.Printf("Actualización automática exitosa: %s", result.Message)
	} else {
		log.Printf("Fallo en actualización automática: %s", result.Message)
	}
}

// GetCacheStats obtiene las estadísticas del caché.
func (s *CurrencyService) GetCacheStats(ctx context.Context) (service.CacheStats, error) {
	return s.cacheService.GetStats(ctx)
}
