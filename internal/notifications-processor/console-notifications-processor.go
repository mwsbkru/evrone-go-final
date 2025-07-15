package notifications_processor

import (
	"evrone_course_final/internal/entity"
	"fmt"
)

type ConsoleNotificationsProcessor struct {
	Name string
}

func (c *ConsoleNotificationsProcessor) Process(notification entity.Notification) error {
	// TODO: switch to slog
	fmt.Println("-------console notifications processor-------------")
	fmt.Println(c.Name)
	fmt.Println(notification.Body)
	fmt.Println("---------------------------------------------------")

	if notification.Body == "never" {
		fmt.Println("-------console notifications processor error-------------")
		fmt.Println(c.Name)
		fmt.Println(notification.Body)
		fmt.Println("---------------------------------------------------")
		return fmt.Errorf("raising error always")
	}

	if notification.Body == "retry" && notification.CurrentRetry < 3 {
		fmt.Println("-------console notifications processor retry-------------")
		fmt.Println(c.Name)
		fmt.Println(notification.Body)
		fmt.Println("---------------------------------------------------")
		return fmt.Errorf("raising error on retry")
	}
	return nil
}
