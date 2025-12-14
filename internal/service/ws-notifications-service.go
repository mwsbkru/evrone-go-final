package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mwsbkru/evrone-go-final/internal/entity"

	"github.com/gorilla/websocket"
)

type ReceivedNotificationProcessor func(notification entity.Notification)
type WsConnectionTerminator func(userEmail string)

type WsNotificationsReceiver interface {
	Subscribe(receivedNotificationProcessor ReceivedNotificationProcessor, wsConnectionTerminator WsConnectionTerminator)
	ReceiveNotifications(ctx context.Context, userEmail string)
}

type WsNotificationsService struct {
	connections             map[string]*websocket.Conn
	wsNotificationsReceiver WsNotificationsReceiver
}

func NewWsNotificationsService(wsNotificationsReceiver WsNotificationsReceiver) *WsNotificationsService {
	return &WsNotificationsService{connections: make(map[string]*websocket.Conn), wsNotificationsReceiver: wsNotificationsReceiver}
}

func (u *WsNotificationsService) Run(ctx context.Context) {
	u.wsNotificationsReceiver.Subscribe(u.handleNotification, u.handleConnectionTermination)
}

func (u *WsNotificationsService) HandleConnection(ctx context.Context, userEmail string, connection *websocket.Conn) {
	currentConnection, ok := u.connections[userEmail]
	if ok {
		currentConnection.WriteMessage(websocket.TextMessage, prepareMessageForSending("new attempt to connect to WS, terminating current connection"))
		u.handleConnectionTermination(userEmail)

		connection.WriteMessage(websocket.TextMessage, prepareMessageForSending("terminating current connection, try again"))
		connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection closed by server"))
		connection.Close()
	}

	u.connections[userEmail] = connection
	go u.handleConnection(ctx, userEmail, connection)
}

func (u *WsNotificationsService) handleConnection(ctx context.Context, userEmail string, connection *websocket.Conn) {
	slog.Info("New WS connection", slog.String("user_email", userEmail))
	defer slog.Info("WS connection closed", slog.String("user_email", userEmail))
	ctx, cancel := context.WithCancel(ctx)
	go u.wsNotificationsReceiver.ReceiveNotifications(ctx, userEmail)
	u.handleConnectionClosedByUser(userEmail, cancel)
}

func (u *WsNotificationsService) handleConnectionClosedByUser(userEmail string, cancel context.CancelFunc) {
	for {
		slog.Info("Waiting for reading message from WS connection", slog.String("user_email", userEmail))

		if conn, ok := u.connections[userEmail]; ok {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				slog.Error("Error in handleConnectionClosedByUser", slog.String("user_email", userEmail), slog.String("error", err.Error()))
				cancel()
				return
			}

			if messageType == websocket.CloseMessage {
				slog.Info("WS connection closed by user", slog.String("user_email", userEmail))
				cancel()
				return
			}
		} else {
			return
		}
	}
}

func (u *WsNotificationsService) handleNotification(notification entity.Notification) {
	if conn, ok := u.connections[notification.UserEmail]; ok {
		conn.WriteMessage(websocket.TextMessage, prepareMessageForSending(notification.Body))
	}
}

func (u *WsNotificationsService) handleConnectionTermination(userEmail string) {
	slog.Info("handleConnectionTermination run", slog.String("user_email", userEmail))
	u.terminateConnection(userEmail)
}

func (u *WsNotificationsService) terminateConnection(userEmail string) {
	slog.Info("Termination connection", slog.String("user_email", userEmail))
	if conn, ok := u.connections[userEmail]; ok {
		delete(u.connections, userEmail)
		conn.WriteMessage(websocket.TextMessage, prepareMessageForSending("connection closed by server"))
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection closed by server"))
		conn.Close()
	}
}

func prepareMessageForSending(message string) []byte {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return []byte(fmt.Sprintf("[%s] %s", currentTime, message))
}
