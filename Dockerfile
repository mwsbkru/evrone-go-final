# Базовый образ
FROM golang:1.25-alpine AS builder

# Установка зависимостей
RUN apk update && apk add --no-cache git

# Копирование исходников
WORKDIR /app
COPY . .

ENV GOPROXY=direct
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache

ARG APP_SUBDIR

# Сборка бинарника старая версия с go mod download смотри в истории git
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/${APP_SUBDIR}/main.go

# Финальный образ
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

# Запуск приложения
CMD ["./main"]
