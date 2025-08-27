// Package entity contiene las entidades de dominio del sistema.
package entity

import (
	"time"
)

// Currency representa una moneda con su valor de cambio.
type Currency struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
	Source    string    `json:"source"`
}

// NewCurrency crea una nueva instancia de Currency.
func NewCurrency(id, name string, value float64, source string) *Currency {
	return &Currency{
		ID:        id,
		Name:      name,
		Value:     value,
		UpdatedAt: time.Now(),
		Source:    source,
	}
}

// IsValid verifica si la moneda tiene datos válidos.
func (c *Currency) IsValid() bool {
	return c.ID != "" && c.Name != "" && c.Value > 0
}

// IsStale verifica si la información de la moneda está desactualizada.
func (c *Currency) IsStale(maxAge time.Duration) bool {
	return time.Since(c.UpdatedAt) > maxAge
}
