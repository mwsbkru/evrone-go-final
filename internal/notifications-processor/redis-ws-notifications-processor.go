package notifications_processor

import (
	"context"
	"encoding/json"
	"errors"
	"evrone_course_final/internal/entity"
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

func (r *RedisWSNotificationsProcessor) Process(notification *entity.Notification) error {
	return errors.New("boom!!!")
	// Проверяем, что уведомление не nil
	if notification == nil {
		return errors.New("notification cannot be nil")
	}

	// Преобразуем структуру в JSON
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return reportAndWrapErrorWs(err, notification.CurrentRetry)
	}

	// TODO: пробросить сюда контекст
	ctx := context.Background()

	// Формируем имя потока на основе email пользователя
	streamName := fmt.Sprintf("notifications:%s", notification.UserEmail)

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

func reportAndWrapErrorWs(err error, currentRetry int) error {
	slog.Error("EmailNotificationsProcessor error send notification", slog.String("error", err.Error()), slog.Int("current retry", currentRetry))
	return fmt.Errorf("EmailNotificationsProcessor error send notification: %w", err)
}
