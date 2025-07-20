package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

// Config Main config of application
type Config struct {
	Host                              string `env:"HOST" env-default:"0.0.0.0"`
	Port                              string `env:"PORT" env-default:"8080"`
	NotificationsRetryCount           int    `env:"NOTIFICATIONS_RETRY_COUNT" env-default:"3"`
	NotificationsRetryIntervalSeconds int    `env:"NOTIFICATIONS_RETRY_INTERVAL_SECONDS" env-default:"3"`
	KafkaBrokers                      string `env:"KAFKA_BROKERS"`
	KafkaConsumerGroupID              string `env:"KAFKA_CONSUMER_GROUP_ID"  env-default:"notifications-processor-async"`
	KafkaTimeoutSeconds               int    `env:"KAFKA_TIMEOUT_SECONDS" env-default:"6"`
	KafkaIntervalSeconds              int    `env:"KAFKA_INTERVAL_SECONDS" env-default:"3"`
	KafkaTopicEmailNotifications      string `env:"KAFKA_TOPIC_EMAIL_NOTIFICATIONS"`
	KafkaTopicPushNotifications       string `env:"KAFKA_TOPIC_PUSH_NOTIFICATIONS"`
	KafkaTopicWSNotifications         string `env:"KAFKA_TOPIC_WS_NOTIFICATIONS"`
	RedisAddr                         string `env:"REDIS_ADDR"`
	RedisDB                           int    `env:"REDIS_DB"`
	RedisMaxRetries                   int    `env:"HOST" env-default:"5"`
	RedisTimeoutSeconds               int    `env:"HOST" env-default:"5"`
	SmtpServerHost                    string `env:"SMTP_SERVER_HOST"`
	SmtpServerPort                    int    `env:"SMTP_SERVER_PORT"`
	SmtpTimeoutSeconds                int    `env:"SMTP_TIMEOUT_SECONDS" env-default:"10"`
	SmtpUsername                      string `env:"SMTP_USERNAME" env-default:""`
	SmtpPassword                      string `env:"SMTP_PASSWORD" env-default:""`
	FromEmail                         string `env:"FROM_EMAIL" env-default:"email@notificator.ru"`
}

// NewConfig returns initialized config
func NewConfig() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("Не удалось прочитать параметры конфига: %w", err)
	}

	return &cfg, nil
}
