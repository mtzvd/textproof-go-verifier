# Multi-stage build для минимизации размера образа

# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Установка необходимых инструментов
RUN apk add --no-cache git make

# Установка Templ
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Генерируем темплейты
RUN templ generate

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o textproof ./cmd/server/main.go

# Stage 2: Runtime
FROM alpine:latest

# Устанавливаем CA сертификаты для HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем собранное приложение из builder stage
COPY --from=builder /app/textproof .

# Копируем статические файлы
COPY --from=builder /app/web/static ./web/static

# Создаём директорию для данных
RUN mkdir -p /app/data

# Открываем порт
EXPOSE 8080

# Том для персистентного хранения данных
VOLUME ["/app/data"]

# Запускаем приложение
CMD ["./textproof", "-port", "8080", "-data-dir", "/app/data"]