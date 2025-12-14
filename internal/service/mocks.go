package service

import (
	"context"

	"github.com/mwsbkru/evrone-go-final/internal/entity"
	"github.com/stretchr/testify/mock"
)

// MockNotificationsObserver is a mock implementation of NotificationsObserver
type MockNotificationsObserver struct {
	mock.Mock
}

func (m *MockNotificationsObserver) Subscribe(subscriber NotificationsSubscriber, terminator Terminator) {
	m.Called(subscriber, terminator)
}

func (m *MockNotificationsObserver) StartListening(ctx context.Context) {
	m.Called(ctx)
}

// MockNotificationsProcessor is a mock implementation of NotificationsProcessor
type MockNotificationsProcessor struct {
	mock.Mock
}

func (m *MockNotificationsProcessor) Process(ctx context.Context, notification *entity.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

// MockDeadNotificationsProcessor is a mock implementation of DeadNotificationsProcessor
type MockDeadNotificationsProcessor struct {
	mock.Mock
}

// MockWsNotificationsReceiver is a mock implementation of WsNotificationsReceiver
type MockWsNotificationsReceiver struct {
	mock.Mock
}

func (m *MockWsNotificationsReceiver) Subscribe(receivedNotificationProcessor ReceivedNotificationProcessor, wsConnectionTerminator WsConnectionTerminator) {
	m.Called(receivedNotificationProcessor, wsConnectionTerminator)
}

func (m *MockWsNotificationsReceiver) ReceiveNotifications(ctx context.Context, userEmail string) {
	m.Called(ctx, userEmail)
}
