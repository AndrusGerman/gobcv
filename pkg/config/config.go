// Package config maneja la configuración de la aplicación.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config contiene toda la configuración de la aplicación.
type Config struct {
	Server   ServerConfig   `json:"server"`
	Cache    CacheConfig    `json:"cache"`
	Scraper  ScraperConfig  `json:"scraper"`
	Database DatabaseConfig `json:"database"`
}

// ServerConfig contiene la configuración del servidor HTTP.
type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// CacheConfig contiene la configuración del caché.
type CacheConfig struct {
	DefaultTTL    time.Duration `json:"default_ttl"`
	CleanupPeriod time.Duration `json:"cleanup_period"`
	MaxItems      int           `json:"max_items"`
}

// ScraperConfig contiene la configuración del scraper.
type ScraperConfig struct {
	BaseURL         string        `json:"base_url"`
	Timeout         time.Duration `json:"timeout"`
	RefreshInterval time.Duration `json:"refresh_interval"`
	UserAgent       string        `json:"user_agent"`
}

// DatabaseConfig contiene la configuración de la base de datos (para futuras extensiones).
type DatabaseConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoadConfig carga la configuración desde variables de entorno con valores por defecto.
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvOrDefault("SERVER_PORT", "8080"),
			Host:         getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnvOrDefault("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnvOrDefault("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnvOrDefault("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Cache: CacheConfig{
			DefaultTTL:    getDurationEnvOrDefault("CACHE_DEFAULT_TTL", 5*time.Minute),
			CleanupPeriod: getDurationEnvOrDefault("CACHE_CLEANUP_PERIOD", 5*time.Minute),
			MaxItems:      getIntEnvOrDefault("CACHE_MAX_ITEMS", 1000),
		},
		Scraper: ScraperConfig{
			BaseURL:         getEnvOrDefault("SCRAPER_BASE_URL", "https://www.bcv.org.ve/"),
			Timeout:         getDurationEnvOrDefault("SCRAPER_TIMEOUT", 30*time.Second),
			RefreshInterval: getDurationEnvOrDefault("SCRAPER_REFRESH_INTERVAL", 15*time.Minute),
			UserAgent:       getEnvOrDefault("SCRAPER_USER_AGENT", "BCV-Currency-API/1.0"),
		},
		Database: DatabaseConfig{
			Type:     getEnvOrDefault("DB_TYPE", "memory"),
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			Database: getEnvOrDefault("DB_NAME", "currencies"),
			Username: getEnvOrDefault("DB_USER", ""),
			Password: getEnvOrDefault("DB_PASSWORD", ""),
		},
	}
}

// getEnvOrDefault obtiene una variable de entorno o retorna un valor por defecto.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnvOrDefault obtiene una duración desde una variable de entorno o retorna un valor por defecto.
func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getIntEnvOrDefault obtiene un entero desde una variable de entorno o retorna un valor por defecto.
func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
