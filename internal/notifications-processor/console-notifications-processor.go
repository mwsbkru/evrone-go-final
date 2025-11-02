package notifications_processor

import (
	"context"
	"errors"
	"log/slog"

	"github.com/mwsbkru/evrone-go-final/internal/entity"
)

type ConsoleNotificationsProcessor struct {
	Name string
}

func (c *ConsoleNotificationsProcessor) Process(ctx context.Context, notification *entity.Notification) error {
	slog.Info("-------console notifications processor-------------")
	slog.Info(c.Name)
	slog.Info(notification.Body)
	slog.Info("---------------------------------------------------")
	return errors.New("Bang!")
}
