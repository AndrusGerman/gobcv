// Package scraper implementa los adaptadores para obtener datos externos.
package scraper

import (
	"context"
	"fmt"
	"sync"

	"gobcv/internal/domain/entity"
	"gobcv/internal/domain/service"
)

// CompositeScraper agrupa múltiples scrapers y combina sus resultados.
type CompositeScraper struct {
	scrapers []service.CurrencyScraper
}

// NewCompositeScraper crea una nueva instancia del scraper compuesto.
func NewCompositeScraper(scrapers ...service.CurrencyScraper) service.CurrencyScraper {
	return &CompositeScraper{
		scrapers: scrapers,
	}
}

// ScrapeCurrencies obtiene monedas de todos los scrapers configurados.
func (s *CompositeScraper) ScrapeCurrencies(ctx context.Context) ([]*entity.Currency, error) {
	var (
		wg            sync.WaitGroup
		mu            sync.Mutex
		allCurrencies []*entity.Currency
		errors        []error
	)

	for _, scraper := range s.scrapers {
		wg.Add(1)
		go func(sc service.CurrencyScraper) {
			defer wg.Done()

			currencies, err := sc.ScrapeCurrencies(ctx)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, err)
			} else {
				allCurrencies = append(allCurrencies, currencies...)
			}
		}(scraper)
	}

	wg.Wait()

	if len(allCurrencies) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all scrapers failed: %v", errors)
	}

	return allCurrencies, nil
}

// ScrapeCurrency busca una moneda específica en los scrapers configurados.
func (s *CompositeScraper) ScrapeCurrency(ctx context.Context, currencyID string) (*entity.Currency, error) {
	// Intentamos secuencialmente ya que es una búsqueda específica
	for _, scraper := range s.scrapers {
		if currency, err := scraper.ScrapeCurrency(ctx, currencyID); err == nil {
			return currency, nil
		}
	}
	return nil, fmt.Errorf("currency %s not found in any scraper", currencyID)
}

// IsHealthy verifica si al menos uno de los scrapers está disponible.
func (s *CompositeScraper) IsHealthy(ctx context.Context) error {
	var errors []error
	healthyCount := 0

	for _, scraper := range s.scrapers {
		if err := scraper.IsHealthy(ctx); err == nil {
			healthyCount++
		} else {
			errors = append(errors, err)
		}
	}

	if healthyCount == 0 {
		return fmt.Errorf("no healthy scrapers available: %v", errors)
	}

	return nil
}
