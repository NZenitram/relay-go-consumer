package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
	topic := os.Getenv("KAFKA_TOPIC")
	offsetReset := os.Getenv("KAFKA_OFFSET_RESET")
	serverID, _ := strconv.Atoi(os.Getenv("SOCKETLABS_SERVER_ID"))
	apiKey := os.Getenv("SOCKETLABS_API_KEY")

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
	var counter int

	for _, partition := range partitionList {
		pc, err := consumer.ConsumePartition(topic, partition, config.Consumer.Offsets.Initial)
		if err != nil {
			log.Fatalf("Failed to start consumer for partition %d: %v", partition, err)
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close() // Ensure the partition consumer is closed
			for msg := range pc.Messages() {
				var emailMessage EmailMessage
				err := json.Unmarshal([]byte(msg.Value), &emailMessage)
				if err != nil {
					log.Fatalf("Failed to parse JSON: %v", err)
				}
				counter = 1
				// Round-robin logic to alternate between senders
				if counter%2 == 0 {
					SendEmailWithSocketLabs(serverID, apiKey, emailMessage)
				} else {
					SendEmailWithPostmark(emailMessage)
				}
				counter++
			}
		}(pc) // Consume messages concurrently
	}

	// Wait forever
	<-context.Background().Done()
}

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Custom unmarshaling logic for EmailAddress
func (e *EmailAddress) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal as a simple string
	var emailString string
	if err := json.Unmarshal(data, &emailString); err == nil {
		// Updated regex pattern to capture Friendly Name and Email Address
		r := regexp.MustCompile(`(?i)(?:"?([^"<]*)"?\s*<([^>]+)>|([^<>\s]+@[^<>\s]+))`)
		matches := r.FindStringSubmatch(emailString)
		if len(matches) > 0 {
			e.Name = strings.TrimSpace(matches[1])
			if matches[2] != "" {
				e.Email = strings.TrimSpace(matches[2])
			} else {
				e.Email = strings.TrimSpace(matches[3])
			}
		} else {
			e.Email = emailString
			e.Name = "" // No name available
		}
		return nil
	}

	// Attempt to unmarshal as an object with email and name
	var alias struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	e.Email = alias.Email
	e.Name = alias.Name
	return nil
}

type Attachment struct {
	Name        string `json:"Name"`
	ContentType string `json:"ContentType"`
	Content     string `json:"Content"`
}

type EmailMessage struct {
	From           EmailAddress      `json:"from"`
	To             []EmailAddress    `json:"to"`
	Cc             []string          `json:"cc"`
	Bcc            []string          `json:"bcc"`
	Subject        string            `json:"subject"`
	Body           string            `json:"body"`
	Attachments    []Attachment      `json:"attachments"`
	Headers        map[string]string `json:"headers"`
	AdditionalData map[string]string `json:"additionaldata"`
}
