package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBinanceScraper_ScrapeCurrencies(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/" { // httptest server path is /
			// Note: real implementation uses full URL, test uses base URL of mock server
		}

		// Mock response
		response := `{
			"code": "000000",
			"message": null,
			"data": [
				{"adv": {"price": "50.5"}},
				{"adv": {"price": "51.5"}}
			],
			"success": true
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Initialize scraper with mock URL
	scraper := &BinanceScraper{
		apiURL:     server.URL,
		httpClient: server.Client(),
	}

	// Test ScrapeCurrencies
	currencies, err := scraper.ScrapeCurrencies(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(currencies) != 1 {
		t.Fatalf("Expected 1 currency, got %d", len(currencies))
	}

	usdt := currencies[0]
	if usdt.ID != "USDT" {
		t.Errorf("Expected ID USDT, got %s", usdt.ID)
	}

	// Expected average: (50.5 + 51.5) / 2 = 51.0
	if usdt.Value != 51.0 {
		t.Errorf("Expected value 51.0, got %f", usdt.Value)
	}
}
