# Базовый образ
FROM golang:1.24-alpine AS builder

# Установка зависимостей
RUN apk update && apk add --no-cache git

# Копирование исходников
WORKDIR /app
COPY . .

# Установка зависимостей проекта
RUN go mod download

ARG APP_SUBDIR

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/${APP_SUBDIR}/main.go

# Финальный образ
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

# Запуск приложения
CMD ["./main"]
