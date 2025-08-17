package usecase

import (
	"context"
	"evrone_course_final/internal/tools"
	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type WsNotificationsUseCase struct {
	connections map[string]*websocket.Conn
	client      *redis.Client
}

func NewWsNotificationsUseCase(client *redis.Client) *WsNotificationsUseCase {
	return &WsNotificationsUseCase{connections: make(map[string]*websocket.Conn), client: client}
}

func (u *WsNotificationsUseCase) HandleConnection(ctx context.Context, userEmail string, connection *websocket.Conn) {
	u.connections[userEmail] = connection
	go u.handleConnection(ctx, userEmail, connection)
}

func (u *WsNotificationsUseCase) handleConnection(ctx context.Context, userEmail string, connection *websocket.Conn) {
	defer connection.Close()
	slog.Info("New WS connection", slog.String("user_email", userEmail))
	defer slog.Info("WS connection closed", slog.String("user_email", userEmail))

	for {
		lastID, err := u.client.Get(ctx, tools.GetUserLastReadedNotificationID(userEmail)).Result()
		if err != nil && err != redis.Nil {
			slog.Error("Can`t fetch last processed ID Redis WS notification for user", slog.String("user_email", userEmail), slog.String("redis_message_id", lastID), slog.String("error", err.Error()))
			break
		}

		if lastID == "" {
			lastID = "0-0"
		}

		entries, err := u.client.XRead(ctx,
			&redis.XReadArgs{
				Streams: []string{tools.GetUserStreamName(userEmail), lastID},
			}).Result()
		if err != nil {
			slog.Error("Can`t read WS notifications from Redis for user", slog.String("user_email", userEmail), slog.String("redis_message_id", lastID), slog.String("error", err.Error()))
			break
		}

		for _, entry := range entries {
			for _, message := range entry.Messages {
				rawNotification, ok := message.Values[tools.REDIS_STREAM_NOTIFICATION_FIELD_NAME]
				if !ok {
					slog.Error("Can`t extract readed notification from Redis for user", slog.String("user_email", userEmail), slog.String("redis_message_id", lastID))
					continue
				}

				notificationString, ok := rawNotification.(string)
				if !ok {
					slog.Error("Can`t process notification from Redis to string for user", slog.String("user_email", userEmail), slog.String("redis_message_id", lastID))
					continue
				}

				u.connections[userEmail].WriteMessage(websocket.TextMessage, []byte(notificationString))

				err := u.client.Set(ctx, tools.GetUserLastReadedNotificationID(userEmail), message.ID, 0).Err()
				if err != nil {
					slog.Error("Can`t set last processed ID of WS notifications from Redis for user", slog.String("user_email", userEmail), slog.String("redis_message_id", lastID), slog.String("error", err.Error()))
					continue
				}
			}
		}
		// TODO: add handle close signal
		// TODO: handle disconnect from user
	}
}
