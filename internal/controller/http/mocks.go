package http

import (
	"github.com/mwsbkru/evrone-go-final/config"
	"github.com/mwsbkru/evrone-go-final/internal/service"
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
