package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type NotificationsUseCase struct {
	notificationChannelUseCase []*NotificationsChannelUseCase
}

func NewNotificationsUseCase(notificationChannelUseCase []*NotificationsChannelUseCase) *NotificationsUseCase {
	return &NotificationsUseCase{notificationChannelUseCase: notificationChannelUseCase}
}

func (n *NotificationsUseCase) Run(ctx context.Context) {
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
