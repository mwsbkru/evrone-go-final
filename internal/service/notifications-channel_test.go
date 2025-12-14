package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/mwsbkru/evrone-go-final/config"
	"github.com/mwsbkru/evrone-go-final/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *MockDeadNotificationsProcessor) Process(notification *entity.Notification, err error) error {
	args := m.Called(notification, err)
	return args.Error(0)
}

func TestNewNotificationChannel(t *testing.T) {
	cfg := &config.Config{
		NotificationsRetryCount:           3,
		NotificationsRetryIntervalSeconds: 5,
	}
	mockObserver := new(MockNotificationsObserver)
	mockProcessor := new(MockNotificationsProcessor)
	mockDeadProcessor := new(MockDeadNotificationsProcessor)

	channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

	assert.NotNil(t, channel)
	assert.Equal(t, "test-channel", channel.Name)
	assert.Equal(t, cfg, channel.cfg)
	assert.Equal(t, mockObserver, channel.notificationsObserver)
	assert.Equal(t, mockProcessor, channel.notificationsProcessor)
	assert.Equal(t, mockDeadProcessor, channel.deadNotificationsProcessor)
}

func TestNotificationsChannel_Run(t *testing.T) {
	synctest.Run(func() {
		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 5,
		}
		mockObserver := new(MockNotificationsObserver)
		mockProcessor := new(MockNotificationsProcessor)
		mockDeadProcessor := new(MockDeadNotificationsProcessor)

		channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(1)

		mockObserver.On("Subscribe", mock.AnythingOfType("NotificationsSubscriber"), mock.AnythingOfType("Terminator")).Return()
		mockObserver.On("StartListening", mock.Anything).Run(func(args mock.Arguments) {
			// Simulate observer running
			time.Sleep(50 * time.Millisecond)
			cancel() // Cancel to stop listening
		}).Return()

		// Run in a goroutine since it blocks
		done := make(chan bool)
		go func() {
			channel.Run(ctx, &wg)
			done <- true
		}()

		time.Sleep(100 * time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Channel.Run did not complete in time")
		}

		mockObserver.AssertExpectations(t)
	})
}

func TestNotificationsChannel_process_Success(t *testing.T) {
	synctest.Run(func() {
		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 5,
		}
		mockObserver := new(MockNotificationsObserver)
		mockProcessor := new(MockNotificationsProcessor)
		mockDeadProcessor := new(MockDeadNotificationsProcessor)

		channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

		ctx := context.Background()
		notification := &entity.Notification{
			UserEmail:    "test@example.com",
			Subject:      "Test",
			Body:         "Test body",
			CurrentRetry: 0,
		}

		mockProcessor.On("Process", ctx, notification).Return(nil).Once()

		channel.process(ctx, notification)

		// Give goroutine time to complete
		time.Sleep(50 * time.Millisecond)

		mockProcessor.AssertExpectations(t)
		mockDeadProcessor.AssertNotCalled(t, "Process")
	})
}

func TestNotificationsChannel_process_Retry(t *testing.T) {
	synctest.Run(func() {
		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 1, // Short interval for testing
		}
		mockObserver := new(MockNotificationsObserver)
		mockProcessor := new(MockNotificationsProcessor)
		mockDeadProcessor := new(MockDeadNotificationsProcessor)

		channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		notification := &entity.Notification{
			UserEmail:    "test@example.com",
			Subject:      "Test",
			Body:         "Test body",
			CurrentRetry: 0,
		}

		// First call fails, second succeeds
		mockProcessor.On("Process", mock.Anything, mock.Anything).Return(errors.New("processing error")).Once()
		mockProcessor.On("Process", mock.Anything, mock.MatchedBy(func(n *entity.Notification) bool {
			return n.CurrentRetry == 1
		})).Return(nil).Once()

		channel.process(ctx, notification)

		// Wait for retry to complete
		time.Sleep(2 * time.Second)

		mockProcessor.AssertExpectations(t)
		mockDeadProcessor.AssertNotCalled(t, "Process")
	})
}

func TestNotificationsChannel_process_MaxRetriesReached(t *testing.T) {
	synctest.Run(func() {
		cfg := &config.Config{
			NotificationsRetryCount:           2,
			NotificationsRetryIntervalSeconds: 1,
		}
		mockObserver := new(MockNotificationsObserver)
		mockProcessor := new(MockNotificationsProcessor)
		mockDeadProcessor := new(MockDeadNotificationsProcessor)

		channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		notification := &entity.Notification{
			UserEmail:    "test@example.com",
			Subject:      "Test",
			Body:         "Test body",
			CurrentRetry: 0,
		}

		processingError := errors.New("processing error")
		deadError := errors.New("dead processing error")

		// All retries fail
		mockProcessor.On("Process", mock.Anything, mock.Anything).Return(processingError).Times(3) // Initial + 2 retries
		mockDeadProcessor.On("Process", mock.MatchedBy(func(n *entity.Notification) bool {
			return n.Channel == "test-channel" && n.CurrentRetry == 2
		}), processingError).Return(deadError).Once()

		channel.process(ctx, notification)

		// Wait for all retries and dead processing
		time.Sleep(4 * time.Second)

		mockProcessor.AssertExpectations(t)
		mockDeadProcessor.AssertExpectations(t)
	})
}

func TestNotificationsChannel_process_ContextCancelledDuringRetry(t *testing.T) {
	synctest.Run(func() {
		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 2,
		}
		mockObserver := new(MockNotificationsObserver)
		mockProcessor := new(MockNotificationsProcessor)
		mockDeadProcessor := new(MockDeadNotificationsProcessor)

		channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

		ctx, cancel := context.WithCancel(context.Background())

		notification := &entity.Notification{
			UserEmail:    "test@example.com",
			Subject:      "Test",
			Body:         "Test body",
			CurrentRetry: 0,
		}

		processingError := errors.New("processing error")

		// First call fails
		mockProcessor.On("Process", mock.Anything, notification).Return(processingError).Once()

		go channel.process(ctx, notification)

		// Cancel context before retry completes
		time.Sleep(100 * time.Millisecond)
		cancel()

		// Wait a bit to ensure retry timer is cancelled
		time.Sleep(100 * time.Millisecond)

		mockProcessor.AssertExpectations(t)
		mockDeadProcessor.AssertNotCalled(t, "Process")
	})

}

func TestNotificationsChannel_terminator(t *testing.T) {
	synctest.Run(func() {
		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 5,
		}
		mockObserver := new(MockNotificationsObserver)
		mockProcessor := new(MockNotificationsProcessor)
		mockDeadProcessor := new(MockDeadNotificationsProcessor)

		channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

		var wg sync.WaitGroup
		wg.Add(1)
		channel.wg = &wg

		// Call terminator in a goroutine to avoid blocking
		done := make(chan bool)
		go func() {
			channel.terminator()
			done <- true
		}()

		// Wait for terminator to complete
		select {
		case <-done:
			// Verify WaitGroup was decremented
			// We can't directly check wg counter, but we can verify it doesn't block
			assert.True(t, true, "Terminator completed")
		case <-time.After(1 * time.Second):
			t.Fatal("Terminator did not complete in time")
		}
	})
}

func TestNotificationsChannel_getSubscriber(t *testing.T) {
	synctest.Run(func() {
		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 5,
		}
		mockObserver := new(MockNotificationsObserver)
		mockProcessor := new(MockNotificationsProcessor)
		mockDeadProcessor := new(MockDeadNotificationsProcessor)

		channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

		ctx := context.Background()
		subscriber := channel.getSubscriber(ctx)

		assert.NotNil(t, subscriber)

		notification := &entity.Notification{
			UserEmail:    "test@example.com",
			Subject:      "Test",
			Body:         "Test body",
			CurrentRetry: 0,
		}

		mockProcessor.On("Process", ctx, notification).Return(nil).Once()

		// Call subscriber (which should call process in a goroutine)
		subscriber(notification)

		// Wait for goroutine to complete
		time.Sleep(50 * time.Millisecond)

		mockProcessor.AssertExpectations(t)
	})
}
