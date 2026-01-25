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

		mockObserver1 := NewMockNotificationsObserver(ctrl)
		mockObserver2 := NewMockNotificationsObserver(ctrl)
		mockObserver3 := NewMockNotificationsObserver(ctrl)

		channels := []*NotificationsChannel{
			NewNotificationChannel(nil, "channel1", mockObserver1, nil, nil),
			NewNotificationChannel(nil, "channel2", mockObserver2, nil, nil),
			NewNotificationChannel(nil, "channel3", mockObserver3, nil, nil),
		}
		service := NewNotificationsService(channels)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		var terminator1, terminator2, terminator3 Terminator

		mockObserver1.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Do(func(subscriber NotificationsSubscriber, term Terminator) {
			terminator1 = term
		}).Times(1)

		mockObserver1.EXPECT().StartListening(gomock.Any()).Do(func(ctx context.Context) {
			<-ctx.Done()
			terminator1()
		}).Times(1)

		mockObserver2.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Do(func(subscriber NotificationsSubscriber, term Terminator) {
			terminator2 = term
		}).Times(1)

		mockObserver2.EXPECT().StartListening(gomock.Any()).Do(func(ctx context.Context) {
			<-ctx.Done()
			terminator2()
		}).Times(1)

		mockObserver3.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Do(func(subscriber NotificationsSubscriber, term Terminator) {
			terminator3 = term
		}).Times(1)

		mockObserver3.EXPECT().StartListening(gomock.Any()).Do(func(ctx context.Context) {
			<-ctx.Done()
			terminator3()
		}).Times(1)

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
