package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

func main() {
	// Конфигурация подключения

	brokers := []string{os.Getenv("KAFKA_BROKERS")}
	topic := "test-topic"
	groupID := "consumer-group"

	// Создание конфигурации консьюмера
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Session.Timeout = 15 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Создание клиента
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	fmt.Println("Created client")

	// Создание консьюмера
	consumer, err := sarama.NewConsumerGroupFromClient(groupID, client)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()
	fmt.Println("Created consumer")

	// Обработка сигналов для корректного завершения
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Start consuming")
	// Основной цикл консьюмера
	for {
		// Запуск цикла обработки сообщений
		err := consumer.Consume(context.Background(), []string{topic}, &handler{})
		if err != nil {
			log.Printf("Error consuming messages: %v", err)
		}

		select {
		case <-signals:
			return
		case <-time.After(250 * time.Millisecond):
			fmt.Println("Chilling out")
		}
	}
}

// Handler реализует интерфейс ConsumerGroupHandler
type handler struct {
	sarama.ConsumerGroupHandler
}

func (h *handler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Обработка сообщений
	for message := range claim.Messages() {
		fmt.Printf("Message received: %s\n", string(message.Value))
		session.MarkMessage(message, "")
	}
	return nil
}
