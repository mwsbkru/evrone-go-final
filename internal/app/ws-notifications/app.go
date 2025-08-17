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
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

func Run(ctx context.Context, cfg *config.Config) {
	kafkaClient, err := tools.PrepareKafkaClient(cfg)
	if err != nil {
		slog.Error("Can`t init Kafka client", slog.String("error", err.Error()))
		return
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   cfg.RedisDB,
	})

	consumerWs, err := sarama.NewConsumerGroupFromClient(cfg.KafkaConsumerGroupID, kafkaClient)
	if err != nil {
		slog.Error("Can`t init Kafka WebSocket", slog.String("error", err.Error()))
		return
	}

	err = tools.EnsureTopicExists(cfg.KafkaTopicDeadNotifications, kafkaClient)
	if err != nil {
		slog.Error("Can`t prepare topic for dead notifications", slog.String("error", err.Error()))
		return
	}

	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer([]string{cfg.KafkaBrokers}, producerConfig)
	if err != nil {
		slog.Error("Can`t init Kafka WebSocket", slog.String("error", err.Error()))
		return
	}

	topicWsNotifications := cfg.KafkaTopicWSNotifications
	kafkaObserverWs := notifications_observer.NewKafkaNotificationsObserver(topicWsNotifications, cfg, consumerWs)

	processorWs := notifications_processor.NewRedisWSNotificationsProcessor(redisClient)
	deadProcessorWs := dead_notifications_processor.NewKafkaDeadNotificationsProcessor(producer, cfg)
	notificationsChannelWs := usecase.NewNotificationChannelUseCase(cfg, "WS processor", kafkaObserverWs, processorWs, deadProcessorWs)

	notificationsChannels := []*usecase.NotificationsChannelUseCase{notificationsChannelWs}
	notificationsUseCase := usecase.NewNotificationsUseCase(notificationsChannels)
	go notificationsUseCase.Run(ctx)

	slog.Info("Starting http server...")
	wsNotificationsUseCase := usecase.NewWsNotificationsUseCase(redisClient)
	server := http.NewServer(cfg, wsNotificationsUseCase)
	http.Serve(server, cfg)
}
