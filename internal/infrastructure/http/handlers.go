// Package http implementa los adaptadores HTTP para la API REST.
package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"gobcv/internal/application/command"
	"gobcv/internal/application/query"
)

// Handlers contiene todos los handlers HTTP de la aplicación.
type Handlers struct {
	refreshHandler     *command.RefreshCurrenciesHandler
	getCurrencyHandler *query.GetCurrencyHandler
	getAllHandler      *query.GetAllCurrenciesHandler
}

// NewHandlers crea una nueva instancia de handlers.
func NewHandlers(
	refreshHandler *command.RefreshCurrenciesHandler,
	getCurrencyHandler *query.GetCurrencyHandler,
	getAllHandler *query.GetAllCurrenciesHandler,
) *Handlers {
	return &Handlers{
		refreshHandler:     refreshHandler,
		getCurrencyHandler: getCurrencyHandler,
		getAllHandler:      getAllHandler,
	}
}

// APIResponse representa la estructura estándar de respuesta de la API.
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// HealthCheck maneja el endpoint de verificación de salud.
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success:   true,
		Message:   "API is running",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RefreshCurrencies maneja el endpoint para refrescar monedas.
func (h *Handlers) RefreshCurrencies(w http.ResponseWriter, r *http.Request) {
	forceRefresh := r.URL.Query().Get("force") == "true"

	cmd := command.RefreshCurrenciesCommand{
		ForceRefresh: forceRefresh,
	}

	result, err := h.refreshHandler.Handle(r.Context(), cmd)

	response := APIResponse{
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()
		response.Message = "Error refreshing currencies"
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = result.Success
		response.Message = result.Message
		response.Data = result
		if !result.Success {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrency maneja el endpoint para obtener una moneda específica.
func (h *Handlers) GetCurrency(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	currencyID := vars["id"]

	useCache := r.URL.Query().Get("cache") != "false" // Por defecto usa caché

	query := query.GetCurrencyQuery{
		CurrencyID: currencyID,
		UseCache:   useCache,
	}

	result, err := h.getCurrencyHandler.Handle(r.Context(), query)

	response := APIResponse{
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()
		response.Message = "Error getting currency"
		w.WriteHeader(http.StatusInternalServerError)
	} else if !result.Success {
		response.Success = false
		response.Message = result.Message
		w.WriteHeader(http.StatusNotFound)
	} else {
		response.Success = true
		response.Message = result.Message
		response.Data = map[string]interface{}{
			"currency":   result.Currency,
			"from_cache": result.FromCache,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllCurrencies maneja el endpoint para obtener todas las monedas.
func (h *Handlers) GetAllCurrencies(w http.ResponseWriter, r *http.Request) {
	useCache := r.URL.Query().Get("cache") != "false" // Por defecto usa caché
	includeStale := r.URL.Query().Get("include_stale") == "true"

	query := query.GetAllCurrenciesQuery{
		UseCache:     useCache,
		IncludeStale: includeStale,
	}

	result, err := h.getAllHandler.Handle(r.Context(), query)

	response := APIResponse{
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()
		response.Message = "Error getting currencies"
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.Message = result.Message
		response.Data = map[string]interface{}{
			"currencies": result.Currencies,
			"count":      result.Count,
			"from_cache": result.FromCache,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCacheStats maneja el endpoint para obtener estadísticas del caché.
func (h *Handlers) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	// Este endpoint requiere acceso directo al servicio de caché
	// Por simplicidad, lo implementaremos más adelante en el servicio de aplicación
	response := APIResponse{
		Success:   true,
		Message:   "Cache stats endpoint - to be implemented",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CORS middleware para permitir requests desde el frontend.
func (h *Handlers) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware registra todas las requests HTTP.
func (h *Handlers) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Crear un ResponseWriter que capture el código de estado
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Log básico - en una aplicación real usarías un logger estructurado
		status := rw.statusCode
		method := r.Method
		path := r.URL.Path

		// Solo log si no es un health check para evitar spam
		if path != "/health" {
			println("HTTP", method, path, status, duration.String())
		}
	})
}

// responseWriter envuelve http.ResponseWriter para capturar el código de estado.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
