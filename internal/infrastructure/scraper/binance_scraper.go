// Package scraper implementa los adaptadores para obtener datos externos.
package scraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gobcv/internal/domain/entity"
	"gobcv/internal/domain/service"
)

// BinanceScraper implementa el servicio de scraping para Binance P2P.
type BinanceScraper struct {
	apiURL     string
	httpClient *http.Client
}

// NewBinanceScraper crea una nueva instancia del scraper de Binance.
func NewBinanceScraper() service.CurrencyScraper {
	return &BinanceScraper{
		apiURL: "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BinanceRequest representa el cuerpo de la petición a la API de Binance.
type BinanceRequest struct {
	Page      int    `json:"page"`
	Rows      int    `json:"rows"`
	Asset     string `json:"asset"`
	TradeType string `json:"tradeType"`
	Fiat      string `json:"fiat"`
}

// BinanceResponse representa la respuesta de la API de Binance.
type BinanceResponse struct {
	Code          string        `json:"code"`
	Message       string        `json:"message"`
	MessageDetail interface{}   `json:"messageDetail"`
	Data          []BinanceData `json:"data"`
	Success       bool          `json:"success"`
}

// BinanceData representa un item de datos en la respuesta.
type BinanceData struct {
	Adv BinanceAdv `json:"adv"`
}

// BinanceAdv representa los detalles del anuncio.
type BinanceAdv struct {
	Price string `json:"price"`
}

// ScrapeCurrencies obtiene las monedas más recientes desde Binance.
func (s *BinanceScraper) ScrapeCurrencies(ctx context.Context) ([]*entity.Currency, error) {
	// Configurar la petición
	payload := BinanceRequest{
		Page:      1,
		Rows:      20,
		Asset:     "USDT",
		TradeType: "BUY",
		Fiat:      "VES",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Ejecutar la petición
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parsear la respuesta
	var binanceResp BinanceResponse
	if err := json.Unmarshal(body, &binanceResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if !binanceResp.Success || len(binanceResp.Data) == 0 {
		return nil, fmt.Errorf("binance API returned unsuccessful response or no data")
	}

	// Calculate el precio promedio
	var total float64
	var count int

	for _, item := range binanceResp.Data {
		var price float64
		_, err := fmt.Sscanf(item.Adv.Price, "%f", &price)
		if err == nil && price > 0 {
			total += price
			count++
		}
	}

	if count == 0 {
		return nil, fmt.Errorf("no valid prices found")
	}

	averagePrice := total / float64(count)

	// Crear la entidad Currency
	usdt := entity.NewCurrency("USDT", "Tether (Binance)", averagePrice, "https://p2p.binance.com/")

	return []*entity.Currency{usdt}, nil
}

// ScrapeCurrency obtiene una moneda específica.
func (s *BinanceScraper) ScrapeCurrency(ctx context.Context, currencyID string) (*entity.Currency, error) {
	if currencyID != "USDT" {
		return nil, fmt.Errorf("currency %s not supported by Binance scraper", currencyID)
	}

	currencies, err := s.ScrapeCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	if len(currencies) > 0 {
		return currencies[0], nil
	}

	return nil, fmt.Errorf("currency USDT not found")
}

// IsHealthy verifica si el servicio de Binance está disponible.
func (s *BinanceScraper) IsHealthy(ctx context.Context) error {
	// Hacemos una petición simple para verificar
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL,
		bytes.NewBuffer([]byte(`{"page":1,"rows":1,"asset":"USDT","tradeType":"BUY","fiat":"VES"}`)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("binance API health check failed with status %d", resp.StatusCode)
	}

	return nil
}
