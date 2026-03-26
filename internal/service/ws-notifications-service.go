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
		u.handleConnectionTermination(userEmail) //nolint:errcheck
	}
	slog.Info("Preparing new WS connection", slog.String("user_email", userEmail))
	u.clients[userEmail] = NewWsNotificationsClient(connection, userEmail)
	go u.handleConnection(ctx, userEmail, currentClient)
}

// TODO remove WsNotificationsClient
func (u *WsNotificationsService) handleConnection(ctx context.Context, userEmail string, wsClient *WsNotificationsClient) {
	slog.Info("New WS connection", slog.String("user_email", userEmail))
	defer slog.Info("WS connection closed", slog.String("user_email", userEmail))
	ctx, cancel := context.WithCancel(ctx)
	go u.wsNotificationsReceiver.ReceiveNotifications(ctx, userEmail)
	u.handleConnectionClosedByUser(userEmail, cancel)
}

func (u *WsNotificationsService) handleConnectionClosedByUser(userEmail string, cancel context.CancelFunc) {
	for {
		slog.Info("Waiting for reading message from WS connection", slog.String("user_email", userEmail))

		if client, ok := u.clients[userEmail]; ok {
			client.Closed()
		} else {
			return
		}
	}
}

func (u *WsNotificationsService) handleNotification(notification entity.Notification) {
	if client, ok := u.clients[notification.UserEmail]; ok {
		client.SendNotification(prepareMessageForSending(notification.Body)) //nolint:errcheck
	}
}

func (u *WsNotificationsService) handleConnectionTermination(userEmail string) {
	slog.Info("handleConnectionTermination run", slog.String("user_email", userEmail))
	u.terminateConnection(userEmail)
}

func (u *WsNotificationsService) terminateConnection(userEmail string) {
	slog.Info("Termination connection", slog.String("user_email", userEmail))
	if client, ok := u.clients[userEmail]; ok {
		delete(u.clients, userEmail)
		client.SendNotification(prepareMessageForSending("connection closed by server")) //nolint:errcheck
		client.Close()                                                                   //nolint:errcheck
	}
}

func prepareMessageForSending(message string) []byte {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return []byte(fmt.Sprintf("[%s] %s", currentTime, message))
}
