// Package repository define los puertos para el acceso a datos.
package repository

import (
	"context"
	"time"

	"gobcv/internal/domain/entity"
)

// CurrencyRepository define el puerto para el repositorio de monedas.
type CurrencyRepository interface {
	// Save guarda o actualiza una moneda en el repositorio.
	Save(ctx context.Context, currency *entity.Currency) error

	// FindByID busca una moneda por su ID.
	FindByID(ctx context.Context, id string) (*entity.Currency, error)

	// FindAll obtiene todas las monedas disponibles.
	FindAll(ctx context.Context) ([]*entity.Currency, error)

	// Delete elimina una moneda del repositorio.
	Delete(ctx context.Context, id string) error

	// FindByLastUpdate busca monedas actualizadas después de una fecha específica.
	FindByLastUpdate(ctx context.Context, since time.Time) ([]*entity.Currency, error)
}
