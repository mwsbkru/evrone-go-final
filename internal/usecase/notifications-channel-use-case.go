package usecase

import (
	"context"
	"evrone_course_final/internal/entity"
	"log/slog"
	"sync"
)

type NotificationsChannelUseCase struct {
	Name                       string
	wg                         *sync.WaitGroup
	notificationsObserver      NotificationsObserver
	notificationsProcessor     NotificationsProcessor
	deadNotificationsProcessor DeadNotificationsProcessor
}

func NewNotificationChannelUseCase(name string, observer NotificationsObserver, processor NotificationsProcessor, deadProcessor DeadNotificationsProcessor) *NotificationsChannelUseCase {
	return &NotificationsChannelUseCase{Name: name, notificationsObserver: observer, notificationsProcessor: processor, deadNotificationsProcessor: deadProcessor}
}

func (n *NotificationsChannelUseCase) Run(ctx context.Context, wg *sync.WaitGroup) {
	n.wg = wg
	n.notificationsObserver.Subscribe(n.subscriber, n.terminator)
	n.notificationsObserver.StartListening(ctx)
}

func (n *NotificationsChannelUseCase) subscriber(notification entity.Notification) {
	// TODO: Add retry
	err := n.notificationsProcessor.Process(notification)
	if err != nil {
		err = n.deadNotificationsProcessor.Process(notification, err)
		if err != nil {
			slog.Error("Can`t process failed notification", slog.String("error", err.Error()), slog.String("process channel", n.Name))
		}
	}
}

func (n *NotificationsChannelUseCase) terminator() {
	n.wg.Done()
}
