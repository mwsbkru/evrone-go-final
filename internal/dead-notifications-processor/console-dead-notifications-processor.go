package dead_notifications_processor

import (
	"evrone_course_final/internal/entity"
	"fmt"
)

type ConsoleDeadNotificationsProcessor struct {
	Name string
}

func (c *ConsoleDeadNotificationsProcessor) Process(notification entity.Notification, err error) error {
	// TODO: switch to slog
	fmt.Println("-------console dead notifications processor-------------")
	fmt.Println(c.Name)
	fmt.Println(notification.Body)
	fmt.Println(err)
	fmt.Println("---------------------------------------------------")
	return nil
}
