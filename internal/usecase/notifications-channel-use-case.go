package usecase

import (
	"context"
	"evrone_course_final/cmd/config"
	"evrone_course_final/internal/entity"
	"log/slog"
	"sync"
	"time"
)

type NotificationsChannelUseCase struct {
	Name                       string
	wg                         *sync.WaitGroup
	notificationsObserver      NotificationsObserver
	notificationsProcessor     NotificationsProcessor
	deadNotificationsProcessor DeadNotificationsProcessor
	cfg                        *config.Config
}

func NewNotificationChannelUseCase(cfg *config.Config, name string, observer NotificationsObserver, processor NotificationsProcessor, deadProcessor DeadNotificationsProcessor) *NotificationsChannelUseCase {
	return &NotificationsChannelUseCase{cfg: cfg, Name: name, notificationsObserver: observer, notificationsProcessor: processor, deadNotificationsProcessor: deadProcessor}
}

func (n *NotificationsChannelUseCase) Run(ctx context.Context, wg *sync.WaitGroup) {
	n.wg = wg
	n.notificationsObserver.Subscribe(n.subscriber, n.terminator)
	n.notificationsObserver.StartListening(ctx)
}

func (n *NotificationsChannelUseCase) subscriber(ctx context.Context, notification *entity.Notification) {
	go n.process(ctx, notification)
}

func (n *NotificationsChannelUseCase) process(ctx context.Context, notification *entity.Notification) {
	slog.Error("Start process notification", slog.String("process channel", n.Name), slog.Int("Retry number", notification.CurrentRetry))
	err := n.notificationsProcessor.Process(notification)
	if err != nil {
		slog.Error("Can`t process failed notification", slog.String("error", err.Error()), slog.String("process channel", n.Name))
		if notification.CurrentRetry < n.cfg.NotificationsRetryCount {
			timer := time.NewTimer(time.Duration(n.cfg.NotificationsRetryIntervalSeconds) * time.Second)
			select {
			case <-timer.C:
				notification.CurrentRetry += 1
				go n.process(ctx, notification)
			case <-ctx.Done():

			}

		} else {
			err = n.deadNotificationsProcessor.Process(notification, err)
			if err != nil {
				slog.Error("Can`t process dead notification", slog.String("error", err.Error()), slog.String("process channel", n.Name))
			}
		}
	}
}

func (n *NotificationsChannelUseCase) terminator() {
	n.wg.Done()
}
