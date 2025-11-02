package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type NotificationsService struct {
	notificationChannelUseCase []*NotificationsChannel
}

func NewNotificationsService(notificationChannelUseCase []*NotificationsChannel) *NotificationsService {
	return &NotificationsService{notificationChannelUseCase: notificationChannelUseCase}
}

func (n *NotificationsService) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(len(n.notificationChannelUseCase))

	slog.Info("Start notifications processing")
	for _, notificationChannelUseCase := range n.notificationChannelUseCase {
		slog.Info(fmt.Sprintf("Start notifications channel: %s", notificationChannelUseCase.Name))
		go notificationChannelUseCase.Run(ctx, &wg)
	}

	slog.Info("Started notifications processing")
	wg.Wait()
	slog.Info("Terminated notifications processing")
}
