package main

import (
	"context"
	"evrone_course_final/config"
	ws_notifications "evrone_course_final/internal/app/ws-notifications"
	"log/slog"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("Не удалось загрузить конфигурацию приложения", slog.String("error", err.Error()))
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := ws_notifications.Run(ctx, cfg); err != nil {
		slog.Error("Failed to run WebSocket notifications service", slog.String("error", err.Error()))
	}
}
