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
	clients                 map[string]*WsNotificationsClient
	wsNotificationsReceiver WsNotificationsReceiver
}

func NewWsNotificationsService(wsNotificationsReceiver WsNotificationsReceiver) *WsNotificationsService {
	return &WsNotificationsService{clients: make(map[string]*WsNotificationsClient), wsNotificationsReceiver: wsNotificationsReceiver}
}

func (u *WsNotificationsService) Run(ctx context.Context) {
	u.wsNotificationsReceiver.Subscribe(u.handleNotification, u.handleConnectionTermination)
}

func (u *WsNotificationsService) HandleConnection(ctx context.Context, userEmail string, connection *websocket.Conn) {
	currentClient, ok := u.clients[userEmail]
	if ok {
		slog.Info("New attempt to connect to WS, terminating current connection", slog.String("user_email", userEmail))
		currentClient.SendNotification(prepareMessageForSending("new attempt to connect to WS, terminating current connection"))
		currentClient.Close()
		delete(u.clients, userEmail)
	}

	slog.Info("Preparing new WS connection", slog.String("user_email", userEmail))
	newClient := NewWsNotificationsClient(connection, userEmail)
	u.clients[userEmail] = newClient
	go u.processConnection(ctx, userEmail)
}

func (u *WsNotificationsService) processConnection(ctx context.Context, userEmail string) {
	slog.Info("New WS connection", slog.String("user_email", userEmail))
	defer slog.Info("WS connection closed", slog.String("user_email", userEmail))
	ctx, cancel := context.WithCancel(ctx)
	go u.wsNotificationsReceiver.ReceiveNotifications(ctx, userEmail)
	u.handleConnectionClosedByUser(userEmail, cancel)
}

func (u *WsNotificationsService) handleNotification(notification entity.Notification) {
	if client, ok := u.clients[notification.UserEmail]; ok {
		defer slog.Info("WS Send notification", slog.String("user_email", notification.UserEmail), slog.String("message", notification.Body))
		client.SendNotification(prepareMessageForSending(notification.Body)) //nolint:errcheck
	} else {
		slog.Info("WS  notification not delivered, user connection not found", slog.String("user_email", notification.UserEmail), slog.String("message", notification.Body))
	}
}

func (u *WsNotificationsService) handleConnectionClosedByUser(userEmail string, cancel context.CancelFunc) {
	if client, ok := u.clients[userEmail]; ok {
		slog.Info("Waiting for closing WS connection by user", slog.String("user_email", userEmail))
		<-client.Closed()
		cancel()
		delete(u.clients, userEmail)
	}
}

func (u *WsNotificationsService) handleConnectionTermination(userEmail string) {
	slog.Info("handleConnectionTermination run", slog.String("user_email", userEmail))
	if client, ok := u.clients[userEmail]; ok {
		slog.Info("Delete client", slog.String("user_email", userEmail))
		delete(u.clients, userEmail)
		client.Close()
	}
}

func prepareMessageForSending(message string) []byte {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return []byte(fmt.Sprintf("[%s] %s", currentTime, message))
}
