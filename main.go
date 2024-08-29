package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	"golang.org/x/exp/rand"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
	emailTopic := os.Getenv("KAFKA_EMAIL_TOPIC")
	webhookTopic := os.Getenv("KAFKA_WEBHOOK_TOPIC")
	offsetReset := os.Getenv("KAFKA_OFFSET_RESET")

	// Set the offset reset policy based on the environment variable
	var offsetResetConfig int64
	if offsetReset == "earliest" {
		offsetResetConfig = sarama.OffsetOldest
	} else {
		offsetResetConfig = sarama.OffsetNewest
	}

	// Create new consumer configuration
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = offsetResetConfig

	// Create new consumer
	consumer, err := sarama.NewConsumer(kafkaBrokers, nil)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Consume messages from the 'emails' topic
	consumeTopic(consumer, emailTopic, config, ProcessEmailMessages)

	// Consume messages from the 'webhook-events' topic
	consumeTopic(consumer, webhookTopic, config, ProcessWebhookMessages)

	// Wait forever
	<-context.Background().Done()
}

func consumeTopic(consumer sarama.Consumer, topic string, config *sarama.Config, processFunc func(*sarama.ConsumerMessage)) {
	partitionList, err := consumer.Partitions(topic)
	if err != nil {
		log.Fatalf("Failed to get the list of partitions for topic %s: %v", topic, err)
	}

	rand.Seed(uint64(time.Now().UnixNano())) // Seed the random number generator
	for _, partition := range partitionList {
		pc, err := consumer.ConsumePartition(topic, partition, config.Consumer.Offsets.Initial)
		if err != nil {
			log.Fatalf("Failed to start consumer for partition %d on topic %s: %v", partition, topic, err)
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close() // Ensure the partition consumer is closed
			for msg := range pc.Messages() {
				processFunc(msg)
			}
		}(pc) // Consume messages concurrently
	}
}
