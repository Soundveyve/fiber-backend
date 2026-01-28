# Многостадийная сборка для уменьшения размера образа
# Стадия 1: Сборка приложения
FROM golang:1.21-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git make

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
# Копируем отдельно для кеширования слоев Docker
COPY go.mod go.sum ./

# Загружаем зависимости
# Этот слой будет закеширован если go.mod/go.sum не изменились
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем приложение
# CGO_ENABLED=0 для статической линковки
# -ldflags="-w -s" уменьшает размер бинарника
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/fiber-backend \
    cmd/api/main.go

# Стадия 2: Финальный образ
FROM alpine:latest

# Устанавливаем CA сертификаты для HTTPS запросов
RUN apk --no-cache add ca-certificates

# Создаем непривилегированного пользователя
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Устанавливаем рабочую директорию
WORKDIR /home/appuser

# Копируем бинарник из builder стадии
COPY --from=builder /app/bin/fiber-backend .

# Меняем владельца файлов
RUN chown -R appuser:appuser /home/appuser

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порт для HTTP API
EXPOSE 3000

# Команда запуска
CMD ["./fiber-backend"]
