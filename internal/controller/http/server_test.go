package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mwsbkru/evrone-go-final/internal/entity/dto"
	"github.com/mwsbkru/evrone-go-final/internal/service"
	"github.com/stretchr/testify/assert"
)

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

func TestNewServer_GetCheckOrigin(t *testing.T) {
	tests := []struct {
		name          string
		checkOrigin   bool
		allowedOrigin string
		requestOrigin string
		expected      bool
	}{
		{
			name:          "WhenCheckOriginFalse",
			checkOrigin:   false,
			allowedOrigin: "http://example.com",
			requestOrigin: "http://malicious.com",
			expected:      true,
		},
		{
			name:          "WhenCheckOriginTrue_AllowedOrigin",
			checkOrigin:   true,
			allowedOrigin: "http://example.com",
			requestOrigin: "http://example.com",
			expected:      true,
		},
		{
			name:          "WhenCheckOriginTrue_DisallowedOrigin",
			checkOrigin:   true,
			allowedOrigin: "http://example.com",
			requestOrigin: "http://malicious.com",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(tt.checkOrigin, tt.allowedOrigin)
			checkOrigin := getCheckOrigin(cfg)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)

			result := checkOrigin(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServer_SubscribeNotifications_InvalidUserEmail(t *testing.T) {
	cfg := createTestConfig(true, "http://example.com")
	mockReceiver := new(service.MockWsNotificationsReceiver)
	wsService := service.NewWsNotificationsService(mockReceiver)
	server := createTestServer(cfg, wsService)

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
		})
	}
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
