package service

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var observerMockFunctionSubscribe = func(args mock.Arguments) {
	// Simulate observer running
	time.Sleep(50 * time.Millisecond)
	terminator := args.Get(1).(Terminator)
	terminator()
}

var observerMockFunctionStartListening = func(args mock.Arguments) {
	ctx := args.Get(0).(context.Context)
	<-ctx.Done()
}

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
		mockObserver := new(MockNotificationsObserver)

		channels := []*NotificationsChannel{
			NewNotificationChannel(nil, "channel1", mockObserver, nil, nil),
			NewNotificationChannel(nil, "channel2", mockObserver, nil, nil),
			NewNotificationChannel(nil, "channel3", mockObserver, nil, nil),
		}
		service := NewNotificationsService(channels)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var terminator Terminator

		mockObserver.On("Subscribe", mock.AnythingOfType("NotificationsSubscriber"), mock.AnythingOfType("Terminator")).Run(func(args mock.Arguments) {
			terminator = args.Get(1).(Terminator)
		}).Return()

		mockObserver.On("StartListening", mock.Anything).Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			<-ctx.Done()
			terminator()
		}).Return()

		done := make(chan bool)
		go func() {
			service.Run(ctx)
			done <- true
		}()

		time.Sleep(100 * time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Service.Run did not complete in time")
		}

		mockObserver.AssertExpectations(t)
	})
}
