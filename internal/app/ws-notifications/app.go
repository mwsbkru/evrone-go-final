package ws_notifications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mwsbkru/evrone-go-final/config"
	"github.com/mwsbkru/evrone-go-final/internal/controller/http"
	dead_notifications_processor "github.com/mwsbkru/evrone-go-final/internal/dead-notifications-processor"
	notifications_observer "github.com/mwsbkru/evrone-go-final/internal/notifications-observer"
	notifications_processor "github.com/mwsbkru/evrone-go-final/internal/notifications-processor"
	"github.com/mwsbkru/evrone-go-final/internal/service"
	"github.com/mwsbkru/evrone-go-final/internal/tools"
	ws_notifications_receivers "github.com/mwsbkru/evrone-go-final/internal/ws-notifications-receivers"

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
		Addr: cfg.Redis.Addr,
		DB:   cfg.Redis.DB,
	})
	defer redisClient.Close()

	consumerWs, err := sarama.NewConsumerGroupFromClient(cfg.Kafka.ConsumerGroupID, kafkaClient)
	if err != nil {
		return fmt.Errorf("can't init Kafka WebSocket consumer: %w", err)
	}
	defer consumerWs.Close()

	err = tools.EnsureTopicExists(cfg.Kafka.TopicDeadNotifications, kafkaClient)
	if err != nil {
		return fmt.Errorf("can't prepare topic for dead notifications: %w", err)
	}

	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer([]string{cfg.Kafka.Brokers}, producerConfig)
	if err != nil {
		return fmt.Errorf("can't init Kafka producer: %w", err)
	}
	defer producer.Close()

	topicWsNotifications := cfg.Kafka.TopicWSNotifications
	kafkaObserverWs := notifications_observer.NewKafkaNotificationsObserver(topicWsNotifications, cfg, consumerWs)

	processorWs := notifications_processor.NewRedisWSNotificationsProcessor(redisClient)
	deadProcessorWs := dead_notifications_processor.NewKafkaDeadNotificationsProcessor(producer, cfg)
	notificationsChannelWs := service.NewNotificationChannelUseCase(cfg, "WS processor", kafkaObserverWs, processorWs, deadProcessorWs)

	notificationsChannels := []*service.NotificationsChannel{notificationsChannelWs}
	notificationsService := service.NewNotificationsService(notificationsChannels)
	go notificationsService.Run(ctx)

	slog.Info("Starting http server...")
	wsNotificationsReceiver := ws_notifications_receivers.NewRedisWsNotificationsReceiver(redisClient, cfg)
	wsNotificationsService := service.NewWsNotificationsService(wsNotificationsReceiver)
	wsNotificationsService.Run(ctx)

	server := http.NewServer(cfg, wsNotificationsService)
	http.Serve(ctx, server, cfg)

	return nil
}
