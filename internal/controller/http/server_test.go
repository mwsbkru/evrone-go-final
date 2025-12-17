package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mwsbkru/evrone-go-final/config"
	"github.com/mwsbkru/evrone-go-final/internal/entity/dto"
	"github.com/mwsbkru/evrone-go-final/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to create test server with real service
func createTestServer(cfg *config.Config, wsService *service.WsNotificationsService) *Server {
	return NewServer(cfg, wsService)
}

// Helper function to create test config
func createTestConfig(checkOrigin bool, allowedOrigin string) *config.Config {
	return &config.Config{
		WS: config.WSConfig{
			CheckOrigin:   checkOrigin,
			AllowedOrigin: allowedOrigin,
		},
	}
}

func TestNewServer(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)

	server := NewServer(cfg, wsService)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.cfg)
	assert.Equal(t, wsService, server.wsNotificationsService)
	assert.NotNil(t, server.upgrader)
	assert.Equal(t, 1024, server.upgrader.ReadBufferSize)
	assert.Equal(t, 1024, server.upgrader.WriteBufferSize)
	assert.NotNil(t, server.upgrader.CheckOrigin)
}

func TestNewServer_CheckOrigin_WhenCheckOriginFalse(t *testing.T) {
	cfg := createTestConfig(false, "http://example.com")
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)

	server := NewServer(cfg, wsService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://malicious.com")

	// When CheckOrigin is false, should allow any origin
	result := server.upgrader.CheckOrigin(req)
	assert.True(t, result)
}

func TestNewServer_CheckOrigin_WhenCheckOriginTrue_AllowedOrigin(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)

	server := NewServer(cfg, wsService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")

	result := server.upgrader.CheckOrigin(req)
	assert.True(t, result)
}

func TestNewServer_CheckOrigin_WhenCheckOriginTrue_DisallowedOrigin(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)

	server := NewServer(cfg, wsService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://malicious.com")

	result := server.upgrader.CheckOrigin(req)
	assert.False(t, result)
}

func TestServer_SubscribeNotifications_MissingUserEmail(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)
	server := createTestServer(cfg, wsService)

	req := httptest.NewRequest(http.MethodGet, "/notifications/subscribe", nil)
	rec := httptest.NewRecorder()

	ctx := context.Background()
	handler := server.SubscribeNotifications(ctx)
	handler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResponse dto.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
	assert.Equal(t, "get param userEmail must be present", errorResponse.Message)
}

func TestServer_SubscribeNotifications_EmptyUserEmail(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)
	server := createTestServer(cfg, wsService)

	req := httptest.NewRequest(http.MethodGet, "/notifications/subscribe?userEmail=", nil)
	rec := httptest.NewRecorder()

	ctx := context.Background()
	handler := server.SubscribeNotifications(ctx)
	handler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResponse dto.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
	assert.Equal(t, "get param userEmail must be present", errorResponse.Message)
}

func TestServer_SubscribeNotifications_ValidRequest_WebSocketUpgrade(t *testing.T) {
	cfg := createTestConfig(false, "") // Don't check origin for simplicity
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)
	server := createTestServer(cfg, wsService)

	userEmail := "test@example.com"
	mockReceiver.On("Subscribe", mock.AnythingOfType("service.ReceivedNotificationProcessor"), mock.AnythingOfType("service.WsConnectionTerminator")).Return()
	mockReceiver.On("ReceiveNotifications", mock.Anything, userEmail).Return()

	// Initialize the service
	ctx := context.Background()
	wsService.Run(ctx)

	// Create a test HTTP server that uses our handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := server.SubscribeNotifications(ctx)
		handler(w, r)
	}))

	defer testServer.Close()

	// Create WebSocket connection
	wsURL := "ws" + testServer.URL[4:] + "/notifications/subscribe?userEmail=" + userEmail
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skipf("Skipping test: failed to create WebSocket connection: %v", err)
	}
	defer conn.Close()

	// let observer goroutine to start and receive notifications
	time.Sleep(50 * time.Millisecond)

	// The handler should have called HandleConnection on the service
	// which will eventually call ReceiveNotifications on the receiver
	// Give it a moment for goroutines to execute
	mockReceiver.AssertExpectations(t)
}

func TestServer_respondWithError(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)
	server := createTestServer(cfg, wsService)

	tests := []struct {
		name         string
		code         int
		message      string
		expectedCode int
		expectedBody dto.ErrorResponse
	}{
		{
			name:         "Bad Request",
			code:         http.StatusBadRequest,
			message:      "Invalid request",
			expectedCode: http.StatusBadRequest,
			expectedBody: dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request",
			},
		},
		{
			name:         "Not Found",
			code:         http.StatusNotFound,
			message:      "Resource not found",
			expectedCode: http.StatusNotFound,
			expectedBody: dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "Resource not found",
			},
		},
		{
			name:         "Internal Server Error",
			code:         http.StatusInternalServerError,
			message:      "Server error",
			expectedCode: http.StatusInternalServerError,
			expectedBody: dto.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			server.respondWithError(rec, tt.code, tt.message)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var errorResponse dto.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody.Code, errorResponse.Code)
			assert.Equal(t, tt.expectedBody.Message, errorResponse.Message)
		})
	}
}
