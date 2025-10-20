package ws_notifications

import (
	"context"
	"evrone_course_final/config"
	"evrone_course_final/internal/controller/http"
	dead_notifications_processor "evrone_course_final/internal/dead-notifications-processor"
	notifications_observer "evrone_course_final/internal/notifications-observer"
	notifications_processor "evrone_course_final/internal/notifications-processor"
	"evrone_course_final/internal/tools"
	"evrone_course_final/internal/usecase"
	ws_notifications_receivers "evrone_course_final/internal/ws-notifications-receivers"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
)

func Run(ctx context.Context, cfg *config.Config) error {
	kafkaClient, err := tools.PrepareKafkaClient(cfg)
	if err != nil {
		return fmt.Errorf("can't init Kafka client: %w", err)
	}
	defer kafkaClient.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   cfg.RedisDB,
	})
	defer redisClient.Close()

	consumerWs, err := sarama.NewConsumerGroupFromClient(cfg.KafkaConsumerGroupID, kafkaClient)
	if err != nil {
		return fmt.Errorf("can't init Kafka WebSocket consumer: %w", err)
	}
	defer consumerWs.Close()

	err = tools.EnsureTopicExists(cfg.KafkaTopicDeadNotifications, kafkaClient)
	if err != nil {
		return fmt.Errorf("can't prepare topic for dead notifications: %w", err)
	}

	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer([]string{cfg.KafkaBrokers}, producerConfig)
	if err != nil {
		return fmt.Errorf("can't init Kafka producer: %w", err)
	}
	defer producer.Close()

	topicWsNotifications := cfg.KafkaTopicWSNotifications
	kafkaObserverWs := notifications_observer.NewKafkaNotificationsObserver(topicWsNotifications, cfg, consumerWs)

	processorWs := notifications_processor.NewRedisWSNotificationsProcessor(redisClient)
	deadProcessorWs := dead_notifications_processor.NewKafkaDeadNotificationsProcessor(producer, cfg)
	notificationsChannelWs := usecase.NewNotificationChannelUseCase(cfg, "WS processor", kafkaObserverWs, processorWs, deadProcessorWs)

	notificationsChannels := []*usecase.NotificationsChannelUseCase{notificationsChannelWs}
	notificationsUseCase := usecase.NewNotificationsUseCase(notificationsChannels)
	go notificationsUseCase.Run(ctx)

	slog.Info("Starting http server...")
	wsNotificationsReceiver := ws_notifications_receivers.NewRedisWsNotificationsReceiver(redisClient, cfg)
	wsNotificationsUseCase := usecase.NewWsNotificationsUseCase(wsNotificationsReceiver)
	wsNotificationsUseCase.Run(ctx)

	server := http.NewServer(cfg, wsNotificationsUseCase)
	http.Serve(ctx, server, cfg)

	return nil
}
