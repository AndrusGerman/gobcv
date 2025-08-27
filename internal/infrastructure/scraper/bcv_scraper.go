// Package scraper implementa el adaptador para obtener datos del BCV.
package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"gobcv/internal/domain/entity"
	"gobcv/internal/domain/service"
)

// BCVScraper implementa el servicio de scraping del Banco Central de Venezuela.
type BCVScraper struct {
	baseURL    string
	httpClient *http.Client
}

// NewBCVScraper crea una nueva instancia del scraper del BCV.
func NewBCVScraper() service.CurrencyScraper {
	return &BCVScraper{
		baseURL: "https://www.bcv.org.ve/",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ScrapeCurrencies obtiene las monedas más recientes desde la fuente externa.
func (s *BCVScraper) ScrapeCurrencies(ctx context.Context) ([]*entity.Currency, error) {
	htmlStr, err := s.getHTML(ctx, s.baseURL)
	if err != nil {
		return nil, fmt.Errorf("error getting HTML: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	bodyNode := s.findByTag(doc, "body")
	if bodyNode == nil {
		return nil, fmt.Errorf("body element not found")
	}

	var currencies []*entity.Currency

	// Obtener Euro
	if euroValue, err := s.extractCurrencyValue(bodyNode, "euro"); err == nil {
		euro := entity.NewCurrency("EUR", "Euro", euroValue, s.baseURL)
		currencies = append(currencies, euro)
	}

	// Obtener Dólar
	if dolarValue, err := s.extractCurrencyValue(bodyNode, "dolar"); err == nil {
		dolar := entity.NewCurrency("USD", "Dólar Americano", dolarValue, s.baseURL)
		currencies = append(currencies, dolar)
	}

	if len(currencies) == 0 {
		return nil, fmt.Errorf("no currencies found")
	}

	return currencies, nil
}

// ScrapeCurrency obtiene una moneda específica desde la fuente externa.
func (s *BCVScraper) ScrapeCurrency(ctx context.Context, currencyID string) (*entity.Currency, error) {
	currencies, err := s.ScrapeCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	for _, currency := range currencies {
		if currency.ID == currencyID {
			return currency, nil
		}
	}

	return nil, fmt.Errorf("currency %s not found", currencyID)
}

// IsHealthy verifica si el servicio de scraping está disponible.
func (s *BCVScraper) IsHealthy(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "HEAD", s.baseURL, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("BCV website returned status %d", resp.StatusCode)
	}

	return nil
}

// getHTML obtiene el HTML de una URL.
func (s *BCVScraper) getHTML(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// extractCurrencyValue extrae el valor de una moneda específica del HTML.
func (s *BCVScraper) extractCurrencyValue(bodyNode *html.Node, currencyID string) (float64, error) {
	// Buscar el contenedor con el ID de la moneda
	currencyContainer := s.findByID(bodyNode, currencyID)
	if currencyContainer == nil {
		return 0, fmt.Errorf("currency container for %s not found", currencyID)
	}

	// Buscar el elemento strong dentro del contenedor
	strongNode := s.findByTag(currencyContainer, "strong")
	if strongNode == nil {
		return 0, fmt.Errorf("strong element not found for %s", currencyID)
	}

	// Obtener el contenido del texto
	if strongNode.FirstChild == nil {
		return 0, fmt.Errorf("no text content found for %s", currencyID)
	}

	currencyContent := strongNode.FirstChild.Data

	// Limpiar y convertir el valor
	cleanValue := strings.ReplaceAll(strings.TrimSpace(currencyContent), ",", ".")
	value, err := strconv.ParseFloat(cleanValue, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing currency value %s: %w", cleanValue, err)
	}

	return value, nil
}

// findByID busca un elemento por su ID en el árbol DOM.
func (s *BCVScraper) findByID(node *html.Node, id string) *html.Node {
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == "id" && attr.Val == id {
				return node
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := s.findByID(child, id); result != nil {
			return result
		}
	}

	return nil
}

// findByTag busca un elemento por su nombre de tag en el árbol DOM.
func (s *BCVScraper) findByTag(node *html.Node, tagName string) *html.Node {
	if node.Type == html.ElementNode && node.Data == tagName {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := s.findByTag(child, tagName); result != nil {
			return result
		}
	}

	return nil
}
