package service

import (
	"context"
	"testing"
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
	// Simulate observer running
	time.Sleep(50 * time.Millisecond)
	stubChan := make(chan interface{})
	stubChan <- true
	close(stubChan)
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
	// Create mock observers
	mockObserver1 := new(MockNotificationsObserver)
	mockObserver2 := new(MockNotificationsObserver)

	// Create channels with mock observers
	channel1 := NewNotificationChannel(nil, "channel1", mockObserver1, nil, nil)
	channel2 := NewNotificationChannel(nil, "channel2", mockObserver2, nil, nil)

	channels := []*NotificationsChannel{channel1, channel2}
	service := NewNotificationsService(channels)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up expectations
	mockObserver1.On("Subscribe", mock.AnythingOfType("NotificationsSubscriber"), mock.AnythingOfType("Terminator")).Run(observerMockFunctionSubscribe).Return()
	mockObserver1.On("StartListening", mock.Anything).Run(observerMockFunctionStartListening).Return()

	mockObserver2.On("Subscribe", mock.AnythingOfType("NotificationsSubscriber"), mock.AnythingOfType("Terminator")).Run(observerMockFunctionSubscribe).Return()
	mockObserver2.On("StartListening", mock.Anything).Run(observerMockFunctionStartListening).Return()

	// Run in a goroutine since it blocks
	done := make(chan bool)
	go func() {
		service.Run(ctx)
		done <- true
	}()

	// Wait for service to start and then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for service to finish
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Service.Run did not complete in time")
	}

	mockObserver1.AssertExpectations(t)
	mockObserver2.AssertExpectations(t)
}

func TestNotificationsService_Run_MultipleChannels(t *testing.T) {
	mockObserver := new(MockNotificationsObserver)

	channels := []*NotificationsChannel{
		NewNotificationChannel(nil, "channel1", mockObserver, nil, nil),
		NewNotificationChannel(nil, "channel2", mockObserver, nil, nil),
		NewNotificationChannel(nil, "channel3", mockObserver, nil, nil),
	}

	service := NewNotificationsService(channels)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up expectations for all channels
	mockObserver.On("Subscribe", mock.AnythingOfType("NotificationsSubscriber"), mock.AnythingOfType("Terminator")).Run(observerMockFunctionSubscribe).Return().Times(3)
	mockObserver.On("StartListening", mock.Anything).Run(observerMockFunctionStartListening).Return().Times(3)

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
}

func TestNotificationsService_Run_StopByContext(t *testing.T) {
	mockObserver := new(MockNotificationsObserver)

	channel := NewNotificationChannel(nil, "channel1", mockObserver, nil, nil)
	channels := []*NotificationsChannel{channel}
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
}
