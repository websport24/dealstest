# Multi-stage build для оптимизации размера образа
# Используем Go 1.23 для совместимости с go.mod (требует go >= 1.23.9)
FROM golang:1.23-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata

# Создаем пользователя для безопасности
RUN adduser -D -g '' appuser

# Устанавливаем рабочую директорию
WORKDIR /build

# Копируем go mod файлы для кэширования зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение с оптимизациями
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o clickcounter ./cmd/server

# Финальный образ
FROM scratch

# Импортируем пользователя и группу из builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Копируем CA сертификаты для HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем timezone данные
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Копируем бинарник
COPY --from=builder /build/clickcounter /app/clickcounter

# Копируем конфигурационные файлы
COPY --from=builder /build/configs /app/configs

# Копируем миграции
COPY --from=builder /build/migrations /app/migrations

# Устанавливаем пользователя
USER appuser

# Устанавливаем рабочую директорию
WORKDIR /app

# Открываем порт
EXPOSE 8080

# Устанавливаем переменные окружения по умолчанию
ENV CLICKCOUNTER_ENVIRONMENT=production
ENV CLICKCOUNTER_CONFIG_PATH=/app/configs/config.yaml

# Health check отключен для scratch образа (будет проверяться через docker-compose)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запускаем приложение
ENTRYPOINT ["/app/clickcounter"] 