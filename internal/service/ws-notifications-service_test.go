package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mwsbkru/evrone-go-final/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// createTestWebSocketConnection creates a test websocket connection using a test server
func createTestWebSocketConnection(t *testing.T) (*websocket.Conn, func()) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Keep connection alive
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))

	// Connect to the test server
	url := "ws" + server.URL[4:] // Convert http to ws
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		server.Close()
		t.Fatalf("Failed to connect to test server: %v", err)
	}

	cleanup := func() {
		conn.Close()
		server.Close()
	}

	return conn, cleanup
}

func TestNewWsNotificationsService(t *testing.T) {
	mockReceiver := new(MockWsNotificationsReceiver)
	service := NewWsNotificationsService(mockReceiver)

	assert.NotNil(t, service)
	assert.NotNil(t, service.connections)
	assert.Equal(t, 0, len(service.connections))
	assert.Equal(t, mockReceiver, service.wsNotificationsReceiver)
}

func TestWsNotificationsService_Run(t *testing.T) {
	mockReceiver := new(MockWsNotificationsReceiver)
	service := NewWsNotificationsService(mockReceiver)

	mockReceiver.On("Subscribe", mock.AnythingOfType("ReceivedNotificationProcessor"), mock.AnythingOfType("WsConnectionTerminator")).Return()

	ctx := context.Background()
	service.Run(ctx)

	mockReceiver.AssertExpectations(t)
}

func TestWsNotificationsService_HandleConnection_NewConnection(t *testing.T) {
	mockReceiver := new(MockWsNotificationsReceiver)
	service := NewWsNotificationsService(mockReceiver)

	// Create a real websocket connection for testing
	conn, cleanup := createTestWebSocketConnection(t)
	defer cleanup()

	userEmail := "test@example.com"

	mockReceiver.On("Subscribe", mock.AnythingOfType("ReceivedNotificationProcessor"), mock.AnythingOfType("WsConnectionTerminator")).Return()
	mockReceiver.On("ReceiveNotifications", mock.Anything, userEmail).Return()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service.Run(ctx)
	service.HandleConnection(ctx, userEmail, conn)

	// Give goroutines time to execute
	time.Sleep(100 * time.Millisecond)

	assert.Contains(t, service.connections, userEmail)
	assert.Equal(t, conn, service.connections[userEmail])
	mockReceiver.AssertExpectations(t)
}

func TestWsNotificationsService_HandleConnection_ExistingConnection(t *testing.T) {
	mockReceiver := new(MockWsNotificationsReceiver)
	service := NewWsNotificationsService(mockReceiver)
	ctx := context.Background()

	userEmail := "test@example.com"
	oldConn, cleanupOld := createTestWebSocketConnection(t)
	defer cleanupOld()
	newConn, cleanupNew := createTestWebSocketConnection(t)
	defer cleanupNew()

	// Set up existing connection
	service.connections[userEmail] = oldConn

	mockReceiver.On("Subscribe", mock.AnythingOfType("ReceivedNotificationProcessor"), mock.AnythingOfType("WsConnectionTerminator")).Return()
	mockReceiver.On("ReceiveNotifications", mock.Anything, userEmail).Return()

	service.Run(ctx)
	service.HandleConnection(ctx, userEmail, newConn)

	// Give goroutines time to execute
	time.Sleep(100 * time.Millisecond)

	// New connection should replace old one
	assert.Equal(t, newConn, service.connections[userEmail])
	mockReceiver.AssertExpectations(t)
}

func TestWsNotificationsService_handleNotification(t *testing.T) {
	mockReceiver := new(MockWsNotificationsReceiver)
	service := NewWsNotificationsService(mockReceiver)

	userEmail := "test@example.com"
	conn, cleanup := createTestWebSocketConnection(t)
	defer cleanup()
	service.connections[userEmail] = conn

	notification := entity.Notification{
		UserEmail: userEmail,
		Body:      "Test notification",
	}

	// The handleNotification will try to write to the connection
	service.handleNotification(notification)

	// Verify the connection is still there
	assert.Contains(t, service.connections, userEmail)

	// Give time for message to be written
	time.Sleep(50 * time.Millisecond)
}

func TestWsNotificationsService_handleNotification_NoConnection(t *testing.T) {
	mockReceiver := new(MockWsNotificationsReceiver)
	service := NewWsNotificationsService(mockReceiver)

	notification := entity.Notification{
		UserEmail: "nonexistent@example.com",
		Body:      "Test notification",
	}

	// Should not panic or error when connection doesn't exist
	service.handleNotification(notification)
}

func TestWsNotificationsService_handleConnectionTermination(t *testing.T) {
	mockReceiver := new(MockWsNotificationsReceiver)
	service := NewWsNotificationsService(mockReceiver)

	userEmail := "test@example.com"
	conn, cleanup := createTestWebSocketConnection(t)
	defer cleanup()
	service.connections[userEmail] = conn

	service.handleConnectionTermination(userEmail)

	assert.NotContains(t, service.connections, userEmail)
}

func TestWsNotificationsService_terminateConnection_NoConnection(t *testing.T) {
	mockReceiver := new(MockWsNotificationsReceiver)
	service := NewWsNotificationsService(mockReceiver)

	userEmail := "nonexistent@example.com"

	// Should not panic when connection doesn't exist
	service.terminateConnection(userEmail)
}

func TestPrepareMessageForSending(t *testing.T) {
	message := "Test message"
	result := prepareMessageForSending(message)

	assert.NotEmpty(t, result)
	assert.Contains(t, string(result), message)

	// Check format: should start with timestamp in brackets
	assert.Regexp(t, `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] Test message$`, string(result))
}
