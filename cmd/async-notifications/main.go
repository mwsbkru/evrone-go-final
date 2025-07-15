package main

import (
	"context"
	"evrone_course_final/cmd/config"
	async_notifications "evrone_course_final/internal/app/async-notifications"
	"log/slog"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("Не удалось загрузить конфигурацию приложения", slog.String("error", err.Error()))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	async_notifications.Run(ctx, cfg)
}
