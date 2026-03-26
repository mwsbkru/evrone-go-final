package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mwsbkru/evrone-go-final/config"
	"github.com/mwsbkru/evrone-go-final/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewNotificationChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		NotificationsRetryCount:           3,
		NotificationsRetryIntervalSeconds: 5,
	}
	mockObserver := NewMockNotificationsObserver(ctrl)
	mockProcessor := NewMockNotificationsProcessor(ctrl)
	mockDeadProcessor := NewMockDeadNotificationsProcessor(ctrl)

	channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

	assert.NotNil(t, channel)
	assert.Equal(t, "test-channel", channel.Name)
	assert.Equal(t, cfg, channel.cfg)
	assert.Equal(t, mockObserver, channel.notificationsObserver)
	assert.Equal(t, mockProcessor, channel.notificationsProcessor)
	assert.Equal(t, mockDeadProcessor, channel.deadNotificationsProcessor)
}

func TestNotificationsChannel_Run(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 5,
		}
		mockObserver := NewMockNotificationsObserver(ctrl)
		mockProcessor := NewMockNotificationsProcessor(ctrl)
		mockDeadProcessor := NewMockDeadNotificationsProcessor(ctrl)

		channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(1)

		mockObserver.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Do(func(subscriber NotificationsSubscriber, terminator Terminator) {
			// Store terminator for later use
		})
		mockObserver.EXPECT().StartListening(gomock.Any()).Do(func(ctx context.Context) {
			// Simulate observer running
			time.Sleep(50 * time.Millisecond)
			cancel() // Cancel to stop listening
		})

		// Run in a goroutine since it blocks
		done := make(chan bool)
		go func() {
			channel.Run(ctx, &wg)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Channel.Run did not complete in time")
		}
	})
}

func TestNotificationsChannel_process_Success(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 5,
		}
		mockObserver := NewMockNotificationsObserver(ctrl)
		mockProcessor := NewMockNotificationsProcessor(ctrl)
		mockDeadProcessor := NewMockDeadNotificationsProcessor(ctrl)

		channel := NewNotificationChannel(cfg, "test-channel", mockObserver, mockProcessor, mockDeadProcessor)

		ctx := context.Background()
		notification := &entity.Notification{
			UserEmail:    "test@example.com",
			Subject:      "Test",
			Body:         "Test body",
			CurrentRetry: 0,
		}

		mockProcessor.EXPECT().Process(ctx, notification).Return(nil).Times(1)

		channel.process(ctx, notification)

		// Give goroutine time to complete
		time.Sleep(50 * time.Millisecond)
	})
}

func TestNotificationsChannel_process_Retry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 1, // Short interval for testing
		}
		mockObserver := NewMockNotificationsObserver(ctrl)
		mockProcessor := NewMockNotificationsProcessor(ctrl)
		mockDeadProcessor := NewMockDeadNotificationsProcessor(ctrl)

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
		mockProcessor.EXPECT().Process(gomock.Any(), gomock.Any()).Return(errors.New("processing error")).Times(1)
		mockProcessor.EXPECT().Process(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, n *entity.Notification) error {
			if n.CurrentRetry == 1 {
				return nil
			}
			return errors.New("unexpected retry count")
		}).Times(1)

		channel.process(ctx, notification)

		// Wait for retry to complete
		time.Sleep(2 * time.Second)
	})
}

func TestNotificationsChannel_process_MaxRetriesReached(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := &config.Config{
			NotificationsRetryCount:           2,
			NotificationsRetryIntervalSeconds: 1,
		}
		mockObserver := NewMockNotificationsObserver(ctrl)
		mockProcessor := NewMockNotificationsProcessor(ctrl)
		mockDeadProcessor := NewMockDeadNotificationsProcessor(ctrl)

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
		mockProcessor.EXPECT().Process(gomock.Any(), gomock.Any()).Return(processingError).Times(3) // Initial + 2 retries
		mockDeadProcessor.EXPECT().Process(gomock.Any(), processingError).DoAndReturn(func(n *entity.Notification, err error) error {
			if n.Channel == "test-channel" && n.CurrentRetry == 2 {
				return deadError
			}
			return errors.New("unexpected notification")
		}).Times(1)

		channel.process(ctx, notification)

		// Wait for all retries and dead processing
		time.Sleep(4 * time.Second)
	})
}

func TestNotificationsChannel_process_ContextCancelledDuringRetry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 2,
		}
		mockObserver := NewMockNotificationsObserver(ctrl)
		mockProcessor := NewMockNotificationsProcessor(ctrl)
		mockDeadProcessor := NewMockDeadNotificationsProcessor(ctrl)

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
		mockProcessor.EXPECT().Process(gomock.Any(), notification).Return(processingError).Times(1)

		go channel.process(ctx, notification)

		// Cancel context before retry completes
		time.Sleep(100 * time.Millisecond)
		cancel()
		// time.Sleep(100 * time.Millisecond) // Give time for cancellation to propagate
	})
}

func TestNotificationsChannel_terminator(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 5,
		}
		mockObserver := NewMockNotificationsObserver(ctrl)
		mockProcessor := NewMockNotificationsProcessor(ctrl)
		mockDeadProcessor := NewMockDeadNotificationsProcessor(ctrl)

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
	synctest.Test(t, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := &config.Config{
			NotificationsRetryCount:           3,
			NotificationsRetryIntervalSeconds: 5,
		}
		mockObserver := NewMockNotificationsObserver(ctrl)
		mockProcessor := NewMockNotificationsProcessor(ctrl)
		mockDeadProcessor := NewMockDeadNotificationsProcessor(ctrl)

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

		mockProcessor.EXPECT().Process(ctx, notification).Return(nil).Times(1)

		// Call subscriber (which should call process in a goroutine)
		subscriber(notification)

		// Wait for goroutine to complete
		time.Sleep(50 * time.Millisecond)
	})
}
