// Package main es el punto de entrada de la aplicación API.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gobcv/internal/application/service"
	"gobcv/internal/infrastructure/cache"
	httpInfra "gobcv/internal/infrastructure/http"
	"gobcv/internal/infrastructure/scraper"
	"gobcv/pkg/config"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()

	log.Printf("Iniciando BCV Currency API en %s:%s", cfg.Server.Host, cfg.Server.Port)

	// Inicializar dependencias
	cacheService := cache.NewMemoryCache()
	defer cacheService.Close()

	repository := cache.NewMemoryRepository()
	scraperService := scraper.NewBCVScraper()

	// Inicializar servicios de aplicación
	currencyService := service.NewCurrencyService(repository, scraperService, cacheService)

	// Inicializar handlers HTTP
	handlers := httpInfra.NewHandlers(
		currencyService.GetRefreshHandler(),
		currencyService.GetCurrencyHandler(),
		currencyService.GetAllCurrenciesHandler(),
	)

	// Configurar router
	router := httpInfra.SetupRouter(handlers)

	// Configurar servidor HTTP
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Iniciar actualización periódica de monedas en background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go currencyService.StartPeriodicRefresh(ctx, cfg.Scraper.RefreshInterval)

	// Realizar una actualización inicial
	log.Println("Realizando actualización inicial de monedas...")
	go func() {
		refreshCtx, refreshCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer refreshCancel()

		result, err := currencyService.GetRefreshHandler().Handle(refreshCtx,
			struct {
				ForceRefresh bool `json:"force_refresh"`
			}{ForceRefresh: true})

		if err != nil {
			log.Printf("Error en actualización inicial: %v", err)
		} else if result.Success {
			log.Printf("Actualización inicial exitosa: %s", result.Message)
		} else {
			log.Printf("Fallo en actualización inicial: %s", result.Message)
		}
	}()

	// Canal para señales del sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Iniciar servidor en una goroutine
	go func() {
		log.Printf("Servidor iniciado en http://%s:%s", cfg.Server.Host, cfg.Server.Port)
		log.Println("Endpoints disponibles:")
		log.Println("  GET  /                           - Documentación de la API")
		log.Println("  GET  /api/v1/health              - Health check")
		log.Println("  GET  /api/v1/currencies          - Obtener todas las monedas")
		log.Println("  GET  /api/v1/currencies/{id}     - Obtener moneda específica")
		log.Println("  POST /api/v1/currencies/refresh  - Actualizar monedas")
		log.Println("  GET  /api/v1/cache/stats         - Estadísticas del caché")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error iniciando servidor: %v", err)
		}
	}()

	// Esperar señal de terminación
	<-sigChan
	log.Println("Recibida señal de terminación, cerrando servidor...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error durante el cierre del servidor: %v", err)
	}

	// Cancelar contexto para detener actualización periódica
	cancel()

	log.Println("Servidor cerrado exitosamente")
}
