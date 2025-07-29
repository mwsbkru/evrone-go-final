package usecase

import (
	"context"
	"evrone_course_final/config"
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
	slog.Info("Start process notification", slog.String("process channel", n.Name), slog.Int("Retry number", notification.CurrentRetry))
	err := n.notificationsProcessor.Process(notification)
	if err != nil {
		slog.Error("Can`t process notification", slog.String("error", err.Error()), slog.String("process channel", n.Name))
		if notification.CurrentRetry < n.cfg.NotificationsRetryCount {

			timer := time.NewTimer(time.Duration(n.cfg.NotificationsRetryIntervalSeconds) * time.Second)
			select {
			case <-timer.C:
				notification.CurrentRetry += 1
				slog.Info("Run retry", slog.String("error", err.Error()), slog.String("process channel", n.Name), slog.Int("current retry", notification.CurrentRetry))
				go n.process(ctx, notification)
				timer.Stop()
			case <-ctx.Done():
				// TODO переделать использование контекста
				slog.Info("Stop retry by ctx.Done()", slog.String("error", err.Error()), slog.String("process channel", n.Name), slog.Int("current retry", notification.CurrentRetry))
			}
		} else {
			notification.Channel = n.Name
			slog.Info("Run dead notification process", slog.String("error", err.Error()), slog.String("process channel", n.Name), slog.Int("current retry", notification.CurrentRetry))
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
