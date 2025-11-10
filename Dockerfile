# Dockerfile para BCV Currency API
# Utiliza multi-stage build para optimizar el tamaño de la imagen final

# Etapa de build
FROM golang:1.25.4-alpine AS builder

# Instalar certificados SSL y herramientas de build
RUN apk add --no-cache ca-certificates git tzdata

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar la aplicación
# - CGO_ENABLED=0: Build estático sin dependencias de C
# - GOOS=linux: Target Linux
# - -a: Rebuild todos los paquetes
# - -installsuffix cgo: Sufijo para evitar conflictos con versiones CGO
# - -ldflags: Flags del linker para optimización
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o bin/api \
    cmd/api/main.go

# Etapa final - imagen mínima
FROM scratch

# Copiar certificados SSL para requests HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copiar información de timezone
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copiar el binario compilado
COPY --from=builder /app/bin/api /api

# Exponer puerto
EXPOSE 8080

# Variables de entorno por defecto
ENV SERVER_PORT=8080
ENV SERVER_HOST=0.0.0.0
ENV CACHE_DEFAULT_TTL=5m
ENV SCRAPER_REFRESH_INTERVAL=15m
ENV SCRAPER_TIMEOUT=30s

# Etiquetas de metadata
LABEL maintainer="AndrusCodex"
LABEL description="API REST para obtener tipos de cambio del Banco Central de Venezuela"
LABEL version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/AndrusGerman/gobcv"
LABEL org.opencontainers.image.documentation="https://github.com/AndrusGerman/gobcv/blob/main/README.md"
LABEL org.opencontainers.image.licenses="MIT"

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/api", "health"] || exit 1

# Usuario no-root para seguridad (funciona con scratch)
USER 65534:65534

# Punto de entrada
ENTRYPOINT ["/api"]
