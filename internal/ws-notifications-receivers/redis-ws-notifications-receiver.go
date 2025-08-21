package ws_notifications_receivers

import (
	"context"
	"encoding/json"
	"evrone_course_final/config"
	"evrone_course_final/internal/entity"
	"evrone_course_final/internal/tools"
	"evrone_course_final/internal/usecase"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisWsNotificationsReceiver struct {
	redisClient                   *redis.Client
	receivedNotificationProcessor usecase.ReceivedNotificationProcessor
	wsConnectionTerminator        usecase.WsConnectionTerminator
	cfg                           *config.Config
}

func NewRedisWsNotificationsReceiver(redisClient *redis.Client, cfg *config.Config) *RedisWsNotificationsReceiver {
	return &RedisWsNotificationsReceiver{redisClient: redisClient, cfg: cfg}
}

func (r *RedisWsNotificationsReceiver) Subscribe(receivedNotificationProcessor usecase.ReceivedNotificationProcessor, wsConnectionTerminator usecase.WsConnectionTerminator) {
	r.receivedNotificationProcessor = receivedNotificationProcessor
	r.wsConnectionTerminator = wsConnectionTerminator
}

func (r *RedisWsNotificationsReceiver) ReceiveNotifications(ctx context.Context, userEmail string) {
	slog.Info("Start receive notifications from Redis for user", slog.String("user_email", userEmail))
	// TODO Handle connection termination by user
	for {
		slog.Info("Waiting for reading message from redis or close connection", slog.String("user_email", userEmail))
		select {
		case <-ctx.Done():
			slog.Info("Terminating redis listening, context done", slog.String("user_email", userEmail))
			r.wsConnectionTerminator(userEmail)
			return
		default:
			err := r.receiveNotifications(ctx, userEmail)
			if err != nil {
				slog.Error("Error processing notifications from Redis for user", slog.String("user_email", userEmail), slog.String("error", err.Error()))
			}
		}
	}
}

func (r *RedisWsNotificationsReceiver) receiveNotifications(ctx context.Context, userEmail string) error {
	lastID, err := r.redisClient.Get(ctx, tools.GetUserLastReadedNotificationID(userEmail)).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("can`t fetch last processed ID Redis WS notification for user: %w", err)
	}

	if lastID == "" {
		lastID = "0-0"
	}

	entries, err := r.redisClient.XRead(ctx,
		&redis.XReadArgs{
			Streams: []string{tools.GetUserStreamName(userEmail), lastID},
			Block:   time.Duration(r.cfg.RedisTimeoutSeconds) * time.Second,
		}).Result()
	if err != nil {
		return fmt.Errorf("can`t read WS notifications from Redis stream for user: %s; lastID: %s; error: %w", userEmail, lastID, err)
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

			var notification entity.Notification
			err = json.Unmarshal([]byte(notificationString), &notification)
			if err != nil {
				slog.Error("Can`t unmarshal notification from Redis to entity.Notification for user", slog.String("user_email", userEmail), slog.String("redis_message_id", lastID), slog.String("error", err.Error()))
				continue
			}

			r.receivedNotificationProcessor(notification)

			err := r.redisClient.Set(ctx, tools.GetUserLastReadedNotificationID(userEmail), message.ID, 0).Err()
			if err != nil {
				slog.Error("Can`t set last processed ID of WS notifications from Redis for user", slog.String("user_email", userEmail), slog.String("redis_message_id", lastID), slog.String("error", err.Error()))
				continue
			}
		}
	}
	slog.Info("No more notifications to read from Redis for user", slog.String("user_email", userEmail))

	return nil
}
