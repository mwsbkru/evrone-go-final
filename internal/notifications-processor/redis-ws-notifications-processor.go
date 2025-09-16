package notifications_processor

import (
	"context"
	"encoding/json"
	"errors"
	"evrone_course_final/internal/entity"
	"evrone_course_final/internal/tools"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type RedisWSNotificationsProcessor struct {
	client *redis.Client
}

func NewRedisWSNotificationsProcessor(client *redis.Client) *RedisWSNotificationsProcessor {
	return &RedisWSNotificationsProcessor{client: client}
}

func (r *RedisWSNotificationsProcessor) Process(ctx context.Context, notification *entity.Notification) error {
	// Проверяем, что уведомление не nil
	if notification == nil {
		return errors.New("notification cannot be nil")
	}

	// Преобразуем структуру в JSON
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return reportAndWrapErrorWs(err, notification.CurrentRetry)
	}

	// Формируем имя потока на основе email пользователя
	streamName := tools.GetUserStreamName(notification.UserEmail)

	// Добавляем сообщение в поток Redis
	_, err = r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		ID:     "*", // Используем автоинкремент ID
		Values: map[string]interface{}{
			"notification": string(notificationJSON),
		},
	}).Result()

	if err != nil {
		return reportAndWrapErrorWs(err, notification.CurrentRetry)
	}

	return nil
}

func (r *RedisWSNotificationsProcessor) Terminate() {
	slog.Info("Terminating RedisWSNotificationsProcessor")
	r.client.Close()
}

func reportAndWrapErrorWs(err error, currentRetry int) error {
	slog.Error("EmailNotificationsProcessor error send notification", slog.String("error", err.Error()), slog.Int("current retry", currentRetry))
	return fmt.Errorf("EmailNotificationsProcessor error send notification: %w", err)
}
