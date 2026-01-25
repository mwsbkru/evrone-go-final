package service

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewNotificationsService(t *testing.T) {
	channels := []*NotificationsChannel{
		NewNotificationChannel(nil, "channel1", nil, nil, nil),
		NewNotificationChannel(nil, "channel2", nil, nil, nil),
	}

	service := NewNotificationsService(channels)

	assert.NotNil(t, service)
	assert.Equal(t, channels, service.notificationChannel)
	assert.Equal(t, 2, len(service.notificationChannel))
}

func TestNewNotificationsService_EmptyChannels(t *testing.T) {
	channels := []*NotificationsChannel{}
	service := NewNotificationsService(channels)

	assert.NotNil(t, service)
	assert.Equal(t, 0, len(service.notificationChannel))
}

func TestNotificationsService_Run(t *testing.T) {
	synctest.Run(func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockObserver := NewMockNotificationsObserver(ctrl)

		channels := []*NotificationsChannel{
			NewNotificationChannel(nil, "channel1", mockObserver, nil, nil),
			NewNotificationChannel(nil, "channel2", mockObserver, nil, nil),
			NewNotificationChannel(nil, "channel3", mockObserver, nil, nil),
		}
		service := NewNotificationsService(channels)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		var terminator Terminator

		mockObserver.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Do(func(subscriber NotificationsSubscriber, term Terminator) {
			terminator = term
		}).Times(3)

		mockObserver.EXPECT().StartListening(gomock.Any()).Do(func(ctx context.Context) {
			<-ctx.Done()
			terminator()
		}).Times(3)

		done := make(chan bool)
		go func() {
			service.Run(ctx)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Service.Run did not complete in time")
		}
	})
}
