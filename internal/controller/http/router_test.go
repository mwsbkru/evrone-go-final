package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	websocket "github.com/gorilla/websocket"
	"github.com/mwsbkru/evrone-go-final/internal/service"
	"github.com/stretchr/testify/mock"
)

func TestRouter_Serve(t *testing.T) {
	fmt.Println("TestRouter_Serve started")
	cfg := createTestConfig(false, "")
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)
	server := createTestServer(cfg, wsService)

	router := http.NewServeMux()
	ctx := context.Background()
	router.HandleFunc("GET /notifications/subscribe", server.SubscribeNotifications(ctx))

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	userEmail := "test@example.com"

	tests := []struct {
		name                    string
		path                    string
		shouldFail              bool
		shouldCheckExpectations bool
		description             string
	}{
		{
			name:        "exact path without email",
			path:        "/notifications/subscribe",
			shouldFail:  true, // Route found, but userEmail missing
			description: "Exact path should match",
		},
		{
			name:        "path with trailing slash",
			path:        "/notifications/subscribe/",
			shouldFail:  true,
			description: "Path with trailing slash should not match",
		},
		{
			name:        "case sensitive path",
			path:        "/Notifications/Subscribe",
			shouldFail:  true,
			description: "Path should be case sensitive",
		},
		{
			name:        "path with query params but no userEmail",
			path:        "/notifications/subscribe?other=value",
			shouldFail:  true, // Route found, but userEmail missing
			description: "Query params should not affect route matching",
		},
		{
			name:                    "correct path",
			path:                    "/notifications/subscribe?userEmail=" + userEmail,
			shouldFail:              false, // Route found, but userEmail missing
			shouldCheckExpectations: true,
			description:             "Successfully connected to WebSocket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsURL := "ws" + testServer.URL[4:] + tt.path
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			syncChannelReceive := make(chan bool)

			if tt.shouldCheckExpectations {
				mockReceiver.On("ReceiveNotifications", mock.Anything, userEmail).Return().Run(func(args mock.Arguments) {
					syncChannelReceive <- true
				}).Once()
			}

			if tt.shouldFail {
				// We expect the connection to fail
				if err == nil {
					conn.Close()
					t.Fatalf("Expected connection to fail, but it succeeded")
				}
				// Connection failed as expected, nothing more to do
				return
			}
			// We expect the connection to succeed
			if err != nil {
				t.Fatalf("Failed to create WebSocket connection: %v", err)
			}
			defer conn.Close()

			if tt.shouldCheckExpectations {
				<-syncChannelReceive
				mockReceiver.AssertExpectations(t)
			}
		})
	}
	fmt.Println("TestRouter_Serve finished")
}
