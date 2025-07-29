package dead_notifications_processor

import (
	"evrone_course_final/internal/entity"
	"log/slog"
)

type ConsoleDeadNotificationsProcessor struct {
	Name string
}

func (c *ConsoleDeadNotificationsProcessor) Process(notification *entity.Notification, err error) error {
	slog.Info("-------console dead notifications processor-------------")
	slog.Info(c.Name)
	slog.Info(notification.Body)
	slog.Info(err.Error())
	slog.Info("---------------------------------------------------")
	return nil
}

func (c *ConsoleDeadNotificationsProcessor) Terminate() {
	slog.Info("Terminating ConsoleDeadNotificationsProcessor")
}
