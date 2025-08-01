package usecase

import (
	"context"
	"evrone_course_final/internal/entity"
)

type NotificationsSubscriber func(notification *entity.Notification)
type Terminator func() // Когда завершили слушать источник нотификаций уведомляем об этом хозяина

type NotificationsObserver interface {
	Subscribe(subscriber NotificationsSubscriber, terminator Terminator)
	StartListening(ctx context.Context)
}

type NotificationsProcessor interface {
	Process(ctx context.Context, notification *entity.Notification) error
	Terminate()
}

type DeadNotificationsProcessor interface {
	Process(notification *entity.Notification, err error) error
	Terminate()
}
