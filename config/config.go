package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// WSConfig WebSocket configuration
type WSConfig struct {
	CheckOrigin   bool   `env:"WS_CHECK_ORIGIN" env-default:"true"`
	AllowedOrigin string `env:"WS_ALLOWED_ORIGIN"`
}

// KafkaConfig Kafka configuration
type KafkaConfig struct {
	Brokers                 string `env:"KAFKA_BROKERS"`
	ConsumerGroupID         string `env:"KAFKA_CONSUMER_GROUP_ID"  env-default:"notifications-processor-async"`
	TimeoutSeconds          int    `env:"KAFKA_TIMEOUT_SECONDS" env-default:"6"`
	IntervalSeconds         int    `env:"KAFKA_INTERVAL_SECONDS" env-default:"3"`
	TopicEmailNotifications string `env:"KAFKA_TOPIC_EMAIL_NOTIFICATIONS"`
	TopicPushNotifications  string `env:"KAFKA_TOPIC_PUSH_NOTIFICATIONS"`
	TopicWSNotifications    string `env:"KAFKA_TOPIC_WS_NOTIFICATIONS"`
	TopicDeadNotifications  string `env:"KAFKA_TOPIC_DEAD_NOTIFICATIONS"`
}

// RedisConfig Redis configuration
type RedisConfig struct {
	Addr           string `env:"REDIS_ADDR"`
	DB             int    `env:"REDIS_DB"`
	MaxRetries     int    `env:"REDIS_MAX_RETRIES" env-default:"5"`
	TimeoutSeconds int    `env:"REDIS_TIMEOUT_SECONDS" env-default:"5"`
}

// EmailConfig Email/SMTP configuration
type EmailConfig struct {
	SmtpServerHost     string `env:"SMTP_SERVER_HOST"`
	SmtpServerPort     int    `env:"SMTP_SERVER_PORT"`
	SmtpTimeoutSeconds int    `env:"SMTP_TIMEOUT_SECONDS" env-default:"10"`
	SmtpUsername       string `env:"SMTP_USERNAME" env-default:""`
	SmtpPassword       string `env:"SMTP_PASSWORD" env-default:""`
	FromEmail          string `env:"FROM_EMAIL" env-default:"email@notificator.ru"`
}

// Config Main config of application
type Config struct {
	Host                              string `env:"HOST" env-default:"0.0.0.0"`
	Port                              string `env:"PORT" env-default:"8080"`
	NotificationsRetryCount           int    `env:"NOTIFICATIONS_RETRY_COUNT" env-default:"3"`
	NotificationsRetryIntervalSeconds int    `env:"NOTIFICATIONS_RETRY_INTERVAL_SECONDS" env-default:"3"`
	WS                                WSConfig
	Kafka                             KafkaConfig
	Redis                             RedisConfig
	Email                             EmailConfig
}

// NewConfig returns initialized config
func NewConfig() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать параметры конфига: %w", err)
	}

	return &cfg, nil
}
