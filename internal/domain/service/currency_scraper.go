// Package service define los puertos para servicios externos.
package service

import (
	"context"

	"gobcv/internal/domain/entity"
)

// CurrencyScraper define el puerto para el servicio de scraping de monedas.
type CurrencyScraper interface {
	// ScrapeCurrencies obtiene las monedas más recientes desde la fuente externa.
	ScrapeCurrencies(ctx context.Context) ([]*entity.Currency, error)

	// ScrapeCurrency obtiene una moneda específica desde la fuente externa.
	ScrapeCurrency(ctx context.Context, currencyID string) (*entity.Currency, error)

	// IsHealthy verifica si el servicio de scraping está disponible.
	IsHealthy(ctx context.Context) error
}
