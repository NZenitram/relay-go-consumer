package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"relay-go-consumer/database"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
)

var (
	seedFlag         = flag.Bool("seed", false, "Seed the database")
	realTimeSeedFlag = flag.Bool("realtime-seed", false, "Perform real-time seeding")
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	seedFlag := flag.Bool("seed", false, "Seed the database with sample data")
	flag.Parse()

	if *seedFlag {
		database.InitDB()
		db := database.GetDB()
		defer database.CloseDB()

		database.SeedDB(db)

		fmt.Println("Database seeded successfully")
	} else if *realTimeSeedFlag {
		// New real-time seeding logic
		database.InitDB()
		db := database.GetDB()
		defer database.CloseDB()

		err := database.RealTimeSeedDB(db)
		if err != nil {
			fmt.Printf("Error during real-time seeding: %v\n", err)
		} else {
			fmt.Println("Real-time seeding started successfully")
		}
	} else {

		kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
		emailTopic := os.Getenv("KAFKA_EMAIL_TOPIC")
		sendgridWebhookTopic := os.Getenv("WEBHOOK_TOPIC_SENDGRID")
		postmarkWebhookTopic := os.Getenv("WEBHOOK_TOPIC_POSTMARK")
		socketlabsWebhookTopic := os.Getenv("WEBHOOK_TOPIC_SOCKETLABS")
		sparkpostWebhookTopic := os.Getenv("WEBHOOK_TOPIC_SPARKPOST")
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
		config.Consumer.Offsets.AutoCommit.Enable = true
		config.Consumer.Offsets.AutoCommit.Interval = time.Second * 5
		config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

		// Create a WaitGroup to wait for all goroutines
		var wg sync.WaitGroup

		// Consume messages from each topic with a unique consumer group
		topics := []struct {
			topic     string
			group     string
			processor func(*sarama.ConsumerMessage)
		}{
			{emailTopic, "email-group", ProcessEmailMessages},
			{sendgridWebhookTopic, "sendgrid-group", ProcessSendgridEvents},
			{postmarkWebhookTopic, "postmark-group", ProcessPostmarkEvents},
			{socketlabsWebhookTopic, "socketlabs-group", ProcessSocketLabsEvents},
			{sparkpostWebhookTopic, "sparkpost-group", ProcessSparkPostEvents},
		}

		for _, t := range topics {
			wg.Add(1)
			go func(topic, group string, processor func(*sarama.ConsumerMessage)) {
				defer wg.Done()
				consumeTopic(kafkaBrokers, topic, group, config, processor)
			}(t.topic, t.group, t.processor)
		}

		// Wait for all goroutines to finish (which they never will in this case)
		wg.Wait()
	}
}

func consumeTopic(brokers []string, topic string, groupID string, config *sarama.Config, processFunc func(*sarama.ConsumerMessage)) {
	for {
		consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
		if err != nil {
			log.Printf("Error creating consumer group client for topic %s: %v", topic, err)
			time.Sleep(5 * time.Second)
			continue
		}

		handler := consumerGroupHandler{processFunc: processFunc}
		for {
			err := consumer.Consume(context.Background(), []string{topic}, handler)
			if err != nil {
				log.Printf("Error from consumer for topic %s: %v", topic, err)
				break
			}
		}

		err = consumer.Close()
		if err != nil {
			log.Printf("Error closing consumer for topic %s: %v", topic, err)
		}
	}
}

type consumerGroupHandler struct {
	processFunc func(*sarama.ConsumerMessage)
}

func (h consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.processFunc(msg)
		sess.MarkMessage(msg, "")
	}
	return nil
}
