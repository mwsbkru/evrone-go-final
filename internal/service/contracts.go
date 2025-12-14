package service

import (
	"context"

	"github.com/mwsbkru/evrone-go-final/internal/entity"
)

type NotificationsSubscriber func(notification *entity.Notification)
type Terminator func() // Когда завершили слушать источник нотификаций уведомляем об этом хозяина

type NotificationsObserver interface {
	Subscribe(subscriber NotificationsSubscriber, terminator Terminator)
	StartListening(ctx context.Context)
}

type NotificationsProcessor interface {
	Process(ctx context.Context, notification *entity.Notification) error
}

type DeadNotificationsProcessor interface {
	Process(notification *entity.Notification, err error) error
}
