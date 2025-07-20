package notifications_processor

import (
	"evrone_course_final/internal/entity"
	"log/slog"
)

type ConsoleNotificationsProcessor struct {
	Name string
}

func (c *ConsoleNotificationsProcessor) Process(notification *entity.Notification) error {
	slog.Info("-------console notifications processor-------------")
	slog.Info(c.Name)
	slog.Info(notification.Body)
	slog.Info("---------------------------------------------------")
	return nil
}
