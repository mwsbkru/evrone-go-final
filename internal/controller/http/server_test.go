package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mwsbkru/evrone-go-final/internal/entity/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	websocket "github.com/gorilla/websocket"
)

func TestNewServer(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockService := new(MockWsNotificationsService)

	server := NewServer(cfg, mockService)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.cfg)
	assert.Equal(t, mockService, server.wsNotificationsService)
	assert.NotNil(t, server.upgrader)
}

func TestGetCheckOrigin(t *testing.T) {
	tests := []struct {
		name          string
		checkOrigin   bool
		allowedOrigin string
		requestOrigin string
		expected      bool
	}{
		{
			name:          "CheckOrigin disabled - allows any origin",
			checkOrigin:   false,
			allowedOrigin: "http://example.com",
			requestOrigin: "http://malicious.com",
			expected:      true,
		},
		{
			name:          "CheckOrigin enabled - allowed origin",
			checkOrigin:   true,
			allowedOrigin: "http://example.com",
			requestOrigin: "http://example.com",
			expected:      true,
		},
		{
			name:          "CheckOrigin enabled - disallowed origin",
			checkOrigin:   true,
			allowedOrigin: "http://example.com",
			requestOrigin: "http://malicious.com",
			expected:      false,
		},
		{
			name:          "CheckOrigin enabled - no origin header",
			checkOrigin:   true,
			allowedOrigin: "http://example.com",
			requestOrigin: "",
			expected:      false,
		},
		{
			name:          "CheckOrigin enabled - empty allowed origin",
			checkOrigin:   true,
			allowedOrigin: "",
			requestOrigin: "http://example.com",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(tt.checkOrigin, tt.allowedOrigin)
			checkOrigin := getCheckOrigin(cfg)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			result := checkOrigin(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServer_respondWithError(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockService := new(MockWsNotificationsService)
	server := createTestServer(cfg, mockService)

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

func TestServer_SubscribeNotifications_MissingUserEmail(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockService := new(MockWsNotificationsService)
	server := createTestServer(cfg, mockService)

	tests := []struct {
		name    string
		url     string
		message string
	}{
		{
			name:    "Missing userEmail parameter",
			url:     "/notifications/subscribe",
			message: "get param userEmail must be present",
		},
		{
			name:    "Empty userEmail parameter",
			url:     "/notifications/subscribe?userEmail=",
			message: "get param userEmail must be present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			ctx := context.Background()
			handler := server.SubscribeNotifications(ctx)
			handler(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var errorResponse dto.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
			assert.Equal(t, tt.message, errorResponse.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func TestServer_SubscribeNotifications_WebSocketUpgradeError(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockService := new(MockWsNotificationsService)
	mockUpgrader := new(MockWebSocketUpgrader)

	server := &Server{
		cfg:                    cfg,
		wsNotificationsService: mockService,
		upgrader:               mockUpgrader,
	}

	req := httptest.NewRequest(http.MethodGet, "/notifications/subscribe?userEmail=test@example.com", nil)
	rec := httptest.NewRecorder()

	mockUpgrader.On("Upgrade", rec, req, mock.Anything).Return(nil, errors.New("upgrade failed"))

	ctx := context.Background()
	handler := server.SubscribeNotifications(ctx)
	handler(rec, req)

	mockUpgrader.AssertExpectations(t)
	mockService.AssertNotCalled(t, "HandleConnection", mock.Anything, mock.Anything, mock.Anything)
}

func TestServer_SubscribeNotifications_Success(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockService := new(MockWsNotificationsService)
	mockUpgrader := new(MockWebSocketUpgrader)

	server := &Server{
		cfg:                    cfg,
		wsNotificationsService: mockService,
		upgrader:               mockUpgrader,
	}

	req := httptest.NewRequest(http.MethodGet, "/notifications/subscribe?userEmail=test@example.com", nil)
	rec := httptest.NewRecorder()

	testConn := websocket.Conn{}

	mockUpgrader.On("Upgrade", rec, req, mock.Anything).Return(&testConn, nil)
	mockService.On("HandleConnection", mock.Anything, "test@example.com", &testConn).Return()

	ctx := context.Background()
	handler := server.SubscribeNotifications(ctx)
	handler(rec, req)

	mockUpgrader.AssertExpectations(t)
	mockService.AssertExpectations(t)
}
