package notifications_processor

import (
	"context"
	"evrone_course_final/internal/entity"
	"log/slog"
)

type ConsoleNotificationsProcessor struct {
	Name string
}

func (c *ConsoleNotificationsProcessor) Process(ctx context.Context, notification *entity.Notification) error {
	slog.Info("-------console notifications processor-------------")
	slog.Info(c.Name)
	slog.Info(notification.Body)
	slog.Info("---------------------------------------------------")
	return nil
}

func (c *ConsoleNotificationsProcessor) Terminate() {
	slog.Info("Terminating ConsoleNotificationsProcessor")
}
