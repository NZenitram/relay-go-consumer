package main

import (
	"github.com/IBM/sarama"
)

func ProcessWebhookMessages(msg *sarama.ConsumerMessage) {

}

type UniversalEvent struct {
	EventID        string
	Timestamp      string
	Provider       string
	EventType      string
	Recipient      string
	MessageID      string
	Reason         string
	RawReason      string
	AdditionalData map[string]interface{}
}
