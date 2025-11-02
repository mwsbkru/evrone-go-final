package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mwsbkru/evrone-go-final/config"
	async_notifications "github.com/mwsbkru/evrone-go-final/internal/app/async-notifications"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("Не удалось загрузить конфигурацию приложения", slog.String("error", err.Error()))
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := async_notifications.Run(ctx, cfg); err != nil {
		slog.Error("Failed to run application", slog.String("error", err.Error()))
	}
}
