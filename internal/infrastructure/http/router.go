// Package http implementa el router HTTP de la aplicación.
package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRouter configura todas las rutas de la API.
func SetupRouter(handlers *Handlers) *mux.Router {
	router := mux.NewRouter()

	// Aplicar middlewares
	router.Use(handlers.CORSMiddleware)
	router.Use(handlers.LoggingMiddleware)

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Currency endpoints
	api.HandleFunc("/currencies", handlers.GetAllCurrencies).Methods("GET")
	api.HandleFunc("/currencies/{id:[A-Z]{3}}", handlers.GetCurrency).Methods("GET")
	api.HandleFunc("/currencies/refresh", handlers.RefreshCurrencies).Methods("POST")

	// Cache endpoints
	api.HandleFunc("/cache/stats", handlers.GetCacheStats).Methods("GET")

	// Documentación básica en la raíz
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"service": "BCV Currency API",
			"version": "1.0.0",
			"description": "API para obtener tipos de cambio del Banco Central de Venezuela",
			"endpoints": {
				"GET /api/v1/health": "Verificación de salud del servicio",
				"GET /api/v1/currencies": "Obtener todas las monedas",
				"GET /api/v1/currencies/{id}": "Obtener una moneda específica (EUR, USD)",
				"POST /api/v1/currencies/refresh": "Actualizar monedas desde BCV",
				"GET /api/v1/cache/stats": "Estadísticas del caché"
			},
			"parameters": {
				"cache": "false para omitir caché (por defecto: true)",
				"include_stale": "true para incluir monedas obsoletas (por defecto: false)",
				"force": "true para forzar actualización (en refresh, por defecto: false)"
			}
		}`))
	}).Methods("GET")

	return router
}
