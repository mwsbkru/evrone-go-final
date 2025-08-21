package usecase

import (
	"context"
	"evrone_course_final/internal/entity"
	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type ReceivedNotificationProcessor func(notification entity.Notification)
type WsConnectionTerminator func(userEmail string)

type WsNotificationsReceiver interface {
	Subscribe(receivedNotificationProcessor ReceivedNotificationProcessor, wsConnectionTerminator WsConnectionTerminator)
	ReceiveNotifications(ctx context.Context, userEmail string)
}

type WsNotificationsUseCase struct {
	connections             map[string]*websocket.Conn
	wsNotificationsReceiver WsNotificationsReceiver
	client                  *redis.Client
}

func NewWsNotificationsUseCase(client *redis.Client, wsNotificationsReceiver WsNotificationsReceiver) *WsNotificationsUseCase {
	return &WsNotificationsUseCase{connections: make(map[string]*websocket.Conn), client: client, wsNotificationsReceiver: wsNotificationsReceiver}
}

func (u *WsNotificationsUseCase) Run(ctx context.Context) {
	u.wsNotificationsReceiver.Subscribe(u.handleNotification, u.handleConnectionTermination)
}

func (u *WsNotificationsUseCase) HandleConnection(ctx context.Context, userEmail string, connection *websocket.Conn) {
	u.connections[userEmail] = connection
	go u.handleConnection(ctx, userEmail, connection)
}

func (u *WsNotificationsUseCase) handleConnection(ctx context.Context, userEmail string, connection *websocket.Conn) {
	defer connection.Close()
	slog.Info("New WS connection", slog.String("user_email", userEmail))
	defer slog.Info("WS connection closed", slog.String("user_email", userEmail))
	ctx, cancel := context.WithCancel(ctx)
	go u.wsNotificationsReceiver.ReceiveNotifications(ctx, userEmail)
	u.handleConnectionClosed(userEmail, cancel)
}

func (u *WsNotificationsUseCase) handleConnectionClosed(userEmail string, cancel context.CancelFunc) {
	for {
		messageType, _, err := u.connections[userEmail].ReadMessage()
		if err != nil {
			slog.Error("Error reading message for close connection from WS connection", slog.String("user_email", userEmail), slog.String("error", err.Error()))
			cancel()
			return
		}

		if messageType == websocket.CloseMessage {
			slog.Error("WS connection closed by user", slog.String("user_email", userEmail))
			cancel()
			return
		}
	}
}

func (u *WsNotificationsUseCase) handleNotification(notification entity.Notification) {
	u.connections[notification.UserEmail].WriteMessage(websocket.TextMessage, []byte(notification.Body))
}

func (u *WsNotificationsUseCase) handleConnectionTermination(userEmail string) {
	slog.Info("Terminating WS connection", slog.String("user_email", userEmail))
	u.destroyConnection(userEmail)
}

func (u *WsNotificationsUseCase) destroyConnection(userEmail string) {
	u.connections[userEmail].Close()
	delete(u.connections, userEmail)
}
