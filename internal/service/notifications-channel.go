package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/mwsbkru/evrone-go-final/config"
	"github.com/mwsbkru/evrone-go-final/internal/entity"
)

type NotificationsChannel struct {
	Name                       string
	wg                         *sync.WaitGroup
	notificationsObserver      NotificationsObserver
	notificationsProcessor     NotificationsProcessor
	deadNotificationsProcessor DeadNotificationsProcessor
	cfg                        *config.Config
}

func NewNotificationChannelUseCase(cfg *config.Config, name string, observer NotificationsObserver, processor NotificationsProcessor, deadProcessor DeadNotificationsProcessor) *NotificationsChannel {
	return &NotificationsChannel{cfg: cfg, Name: name, notificationsObserver: observer, notificationsProcessor: processor, deadNotificationsProcessor: deadProcessor}
}

func (n *NotificationsChannel) Run(ctx context.Context, wg *sync.WaitGroup) {
	n.wg = wg
	n.notificationsObserver.Subscribe(n.getSubscriber(ctx), n.terminator)
	n.notificationsObserver.StartListening(ctx)
}

func (n *NotificationsChannel) getSubscriber(ctx context.Context) NotificationsSubscriber {
	return func(notification *entity.Notification) {
		go n.process(ctx, notification)
	}
}

func (n *NotificationsChannel) process(ctx context.Context, notification *entity.Notification) {
	slog.Info("Start process notification", slog.String("process channel", n.Name), slog.Int("Retry number", notification.CurrentRetry))
	err := n.notificationsProcessor.Process(ctx, notification)
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

func (n *NotificationsChannel) terminator() {
	n.wg.Done()
}
