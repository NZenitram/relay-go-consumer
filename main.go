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
	topic := os.Getenv("KAFKA_TOPIC")
	offsetReset := os.Getenv("KAFKA_OFFSET_RESET")

	// Set the offset reset policy based on the environment variable
	var offsetResetConfig int64
	if offsetReset == "earliest" {
		offsetResetConfig = sarama.OffsetOldest
	} else {
		offsetResetConfig = sarama.OffsetNewest
	}

	// Create new consumer group configuration
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = offsetResetConfig

	// Create new consumer
	consumer, err := sarama.NewConsumer(kafkaBrokers, nil)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Consume messages from the specified topic
	partitionList, err := consumer.Partitions(topic)
	if err != nil {
		log.Fatalf("Failed to get the list of partitions: %v", err)
	}

	// Initialize a counter for round-robin
	// Define weights for each sender

	rand.Seed(uint64(time.Now().UnixNano())) // Seed the random number generator
	for _, partition := range partitionList {
		pc, err := consumer.ConsumePartition(topic, partition, config.Consumer.Offsets.Initial)
		if err != nil {
			log.Fatalf("Failed to start consumer for partition %d: %v", partition, err)
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close() // Ensure the partition consumer is closed
			for msg := range pc.Messages() {
				ProcessEmailMessages(msg)
			}
		}(pc) // Consume messages concurrently
	}

	// Start the HTTP server
	go StartHTTPServer()

	// Wait forever
	<-context.Background().Done()
}
