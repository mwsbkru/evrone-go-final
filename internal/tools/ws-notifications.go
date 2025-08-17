package tools

import "fmt"

const REDIS_STREAM_NOTIFICATION_FIELD_NAME = "notification"

func GetUserStreamName(userName string) string {
	return fmt.Sprintf("notifications:%s", userName)
}

func GetUserLastReadedNotificationID(userName string) string {
	return fmt.Sprintf("last-readed-notification--%s", userName)
}
