package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type NotificationsService struct {
	notificationChannel []*NotificationsChannel
}

func NewNotificationsService(notificationChannel []*NotificationsChannel) *NotificationsService {
	return &NotificationsService{notificationChannel: notificationChannel}
}

func (n *NotificationsService) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(len(n.notificationChannel))

	slog.Info("Start notifications processing")
	for _, notificationChannel := range n.notificationChannel {
		slog.Info(fmt.Sprintf("Start notifications channel: %s", notificationChannel.Name))
		go notificationChannel.Run(ctx, &wg)
	}

	slog.Info("Started notifications processing")
	wg.Wait()
	slog.Info("Terminated notifications processing")
}
