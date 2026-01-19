package http

import (
	"context"
	"net/http"

	"github.com/mwsbkru/evrone-go-final/config"
	"github.com/stretchr/testify/mock"

	websocket "github.com/gorilla/websocket"
)

// MockWsNotificationsService is a mock implementation of WsNotificationsService
type MockWsNotificationsService struct {
	mock.Mock
}

func (m *MockWsNotificationsService) HandleConnection(ctx context.Context, userEmail string, connection *websocket.Conn) {
	m.Called(ctx, userEmail, connection)
}

// Helper function to create test server with mock service
func createTestServer(cfg *config.Config, wsService WsNotificationsService) *Server {
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

// MockWebSocketUpgrader is a mock implementation of WebSocketUpgrader
type MockWebSocketUpgrader struct {
	mock.Mock
}

func (m *MockWebSocketUpgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {
	args := m.Called(w, r, responseHeader)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*websocket.Conn), args.Error(1)
}
