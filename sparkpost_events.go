package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"relay-go-consumer/database"

	"github.com/IBM/sarama"
)

type SparkPostPayload []struct {
	Msys struct {
		MessageEvent *MessageEvent `json:"message_event,omitempty"`
		TrackEvent   *TrackEvent   `json:"track_event,omitempty"`
	} `json:"msys"`
}

func ProcessSparkPostEvents(msg *sarama.ConsumerMessage) {
	var payload struct {
		Headers SparkPostWebhookHeaders `json:"headers"`
		Body    json.RawMessage         `json:"body"`
	}

	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		fmt.Printf("Failed to unmarshal message: %v\n", err)
		return
	}

	var sparkPostPayload SparkPostPayload
	err = json.Unmarshal(payload.Body, &sparkPostPayload)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return
	}

	for _, event := range sparkPostPayload {
		standardizedEvent := standardizeSparkPostEvent(event)
		err = saveStandardizedSparkPostEvent(standardizedEvent)
		if err != nil {
			fmt.Printf("Error saving standardized event: %v\n", err)
		}
	}
}

func standardizeSparkPostEvent(event struct {
	Msys struct {
		MessageEvent *MessageEvent `json:"message_event,omitempty"`
		TrackEvent   *TrackEvent   `json:"track_event,omitempty"`
	} `json:"msys"`
}) StandardizedEvent {
	var standardEvent StandardizedEvent
	var commonFields *CommonEventFields

	if event.Msys.MessageEvent != nil {
		commonFields = &event.Msys.MessageEvent.CommonEventFields
	} else if event.Msys.TrackEvent != nil {
		commonFields = &event.Msys.TrackEvent.CommonEventFields
	} else {
		return standardEvent // Return empty event if no recognized event type
	}

	standardEvent.MessageID = commonFields.MessageID
	standardEvent.Provider = "sparkpost"
	standardEvent.Processed = true

	timestamp, _ := strconv.ParseInt(commonFields.Timestamp, 10, 64)
	standardEvent.ProcessedTime = timestamp

	switch commonFields.Type {
	case "delivery":
		standardEvent.Delivered = true
		standardEvent.DeliveredTime = &timestamp
	case "bounce":
		standardEvent.Bounce = true
		standardEvent.BounceTime = &timestamp
		standardEvent.BounceType = event.Msys.MessageEvent.BounceClass
		if event.Msys.MessageEvent.BounceClass != "" {
			standardEvent.BounceType = event.Msys.MessageEvent.Reason
		}
	case "delay":
		standardEvent.Deferred = true
		standardEvent.DeferredCount = 1
		standardEvent.LastDeferralTime = &timestamp
	case "open":
		standardEvent.Open = true
		standardEvent.OpenCount = 1
		standardEvent.LastOpenTime = &timestamp
		if event.Msys.TrackEvent != nil && event.Msys.TrackEvent.InitialPixel {
			standardEvent.UniqueOpen = true
			standardEvent.UniqueOpenTime = &timestamp
		}
	case "spam_complaint":
		standardEvent.Dropped = true
		standardEvent.DroppedTime = &timestamp
		standardEvent.DroppedReason = "Spam Complaint"
	}

	return standardEvent
}

func saveStandardizedSparkPostEvent(event StandardizedEvent) error {
	database.InitDB()
	db := database.GetDB()

	// First, try to update an existing record
	stmt, err := db.Prepare(`
        UPDATE events SET
            provider = $2,
            processed = $3,
            processed_time = $4,
            delivered = $5,
            delivered_time = COALESCE($6, delivered_time),
            bounce = $7,
            bounce_type = COALESCE($8, bounce_type),
            bounce_time = COALESCE($9, bounce_time),
            deferred = $10,
            deferred_count = deferred_count + $11,
            last_deferral_time = COALESCE($12, last_deferral_time),
            open = $13,
            open_count = open_count + $14,
            last_open_time = COALESCE($15, last_open_time),
            unique_open = $16,
            unique_open_time = COALESCE($17, unique_open_time),
            dropped = $18,
            dropped_time = COALESCE($19, dropped_time),
            dropped_reason = COALESCE($20, dropped_reason)
        WHERE message_id = $1
        RETURNING message_id
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var updatedMessageID string
	err = stmt.QueryRow(
		event.MessageID,
		event.Provider,
		event.Processed,
		event.ProcessedTime,
		event.Delivered,
		event.DeliveredTime,
		event.Bounce,
		event.BounceType,
		event.BounceTime,
		event.Deferred,
		event.DeferredCount,
		event.LastDeferralTime,
		event.Open,
		event.OpenCount,
		event.LastOpenTime,
		event.UniqueOpen,
		event.UniqueOpenTime,
		event.Dropped,
		event.DroppedTime,
		event.DroppedReason,
	).Scan(&updatedMessageID)

	if err == sql.ErrNoRows {
		// If no existing record was updated, insert a new one
		insertStmt, err := db.Prepare(`
            INSERT INTO events (
                message_id, provider, processed, processed_time, delivered, delivered_time,
                bounce, bounce_type, bounce_time, deferred, deferred_count, last_deferral_time,
                open, open_count, last_open_time, unique_open, unique_open_time,
                dropped, dropped_time, dropped_reason
            ) VALUES (
                $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
            )
        `)
		if err != nil {
			return err
		}
		defer insertStmt.Close()

		_, err = insertStmt.Exec(
			event.MessageID,
			event.Provider,
			event.Processed,
			event.ProcessedTime,
			event.Delivered,
			event.DeliveredTime,
			event.Bounce,
			event.BounceType,
			event.BounceTime,
			event.Deferred,
			event.DeferredCount,
			event.LastDeferralTime,
			event.Open,
			event.OpenCount,
			event.LastOpenTime,
			event.UniqueOpen,
			event.UniqueOpenTime,
			event.Dropped,
			event.DroppedTime,
			event.DroppedReason,
		)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

type GeoIP struct {
	Country    string  `json:"country"`
	Region     string  `json:"region"`
	City       string  `json:"city"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Zip        int     `json:"zip"`
	PostalCode string  `json:"postal_code"`
}

type UserAgentParsed struct {
	AgentFamily  string `json:"agent_family"`
	DeviceBrand  string `json:"device_brand"`
	DeviceFamily string `json:"device_family"`
	OSFamily     string `json:"os_family"`
	OSVersion    string `json:"os_version"`
	IsMobile     bool   `json:"is_mobile"`
	IsProxy      bool   `json:"is_proxy"`
	IsPrefetched bool   `json:"is_prefetched"`
}

type CommonEventFields struct {
	ABTestID              string `json:"ab_test_id"`
	ABTestVersion         string `json:"ab_test_version"`
	AmpEnabled            bool   `json:"amp_enabled"`
	CampaignID            string `json:"campaign_id"`
	ClickTracking         bool   `json:"click_tracking"`
	CustomerID            string `json:"customer_id"`
	DelvMethod            string `json:"delv_method"`
	EventID               string `json:"event_id"`
	FriendlyFrom          string `json:"friendly_from"`
	InitialPixel          bool   `json:"initial_pixel"`
	InjectionTime         string `json:"injection_time"`
	IPAddress             string `json:"ip_address"`
	IPPool                string `json:"ip_pool"`
	MailboxProvider       string `json:"mailbox_provider"`
	MailboxProviderRegion string `json:"mailbox_provider_region"`
	MessageID             string `json:"message_id"`
	MsgFrom               string `json:"msg_from"`
	MsgSize               string `json:"msg_size"`
	NumRetries            string `json:"num_retries"`
	OpenTracking          bool   `json:"open_tracking"`
	QueueTime             string `json:"queue_time"`
	RcptMeta              struct {
		CustomKey string `json:"customKey"`
	} `json:"rcpt_meta"`
	RcptTags        []string `json:"rcpt_tags"`
	RcptTo          string   `json:"rcpt_to"`
	RcptHash        string   `json:"rcpt_hash"`
	RawRcptTo       string   `json:"raw_rcpt_to"`
	RcptType        string   `json:"rcpt_type"`
	RecipientDomain string   `json:"recipient_domain"`
	RoutingDomain   string   `json:"routing_domain"`
	ScheduledTime   string   `json:"scheduled_time"`
	SendingIP       string   `json:"sending_ip"`
	SubaccountID    string   `json:"subaccount_id"`
	Subject         string   `json:"subject"`
	TemplateID      string   `json:"template_id"`
	TemplateVersion string   `json:"template_version"`
	Timestamp       string   `json:"timestamp"`
	Transactional   string   `json:"transactional"`
	TransmissionID  string   `json:"transmission_id"`
	Type            string   `json:"type"`
}

type MessageEvent struct {
	CommonEventFields
	BounceClass  string   `json:"bounce_class,omitempty"`
	ErrorCode    string   `json:"error_code,omitempty"`
	RawReason    string   `json:"raw_reason,omitempty"`
	Reason       string   `json:"reason,omitempty"`
	SMSCoding    string   `json:"sms_coding"`
	SMSDst       string   `json:"sms_dst"`
	SMSDstNpi    string   `json:"sms_dst_npi"`
	SMSDstTon    string   `json:"sms_dst_ton"`
	SMSRemoteids []string `json:"sms_remoteids,omitempty"`
	SMSSegments  int      `json:"sms_segments,omitempty"`
	SMSSrc       string   `json:"sms_src"`
	SMSSrcNpi    string   `json:"sms_src_npi"`
	SMSSrcTon    string   `json:"sms_src_ton"`
	OutboundTLS  string   `json:"outbound_tls"`
	RecvMethod   string   `json:"recv_method"`
}

type TrackEvent struct {
	CommonEventFields
	GeoIP           GeoIP           `json:"geo_ip"`
	TargetLinkName  string          `json:"target_link_name"`
	TargetLinkURL   string          `json:"target_link_url"`
	UserAgent       string          `json:"user_agent"`
	UserAgentParsed UserAgentParsed `json:"user_agent_parsed"`
}

type SparkPostWebhookHeaders struct {
	AcceptEncoding      []string `json:"Accept-Encoding"`
	Authorization       []string `josn:"Authorization"`
	ContentLength       []string `json:"Content-Length"`
	ContentType         []string `json:"Content-Type"`
	UserAgent           []string `json:"User-Agent"`
	XForwardedFor       []string `json:"X-Forwarded-For"`
	XForwardedHost      []string `json:"X-Forwarded-Host"`
	XForwardedProto     []string `json:"X-Forwarded-Proto"`
	XSparkpostSignature []string `json:"X-Sparkpost-Signature"`
}

// // ... (keep all your existing struct definitions for GeoIP, UserAgentParsed, etc.)

// type SparkPostEventUnmarshaler interface {
// 	UnmarshalSparkPostEvent(data []byte, headers SparkPostWebhookHeaders) error
// }

// func (c *CommonEventFields) UnmarshalSparkPostEvent(data []byte, headers SparkPostWebhookHeaders) error {
// 	var payload []struct {
// 		Msys struct {
// 			MessageEvent *CommonEventFields `json:"message_event,omitempty"`
// 			TrackEvent   *CommonEventFields `json:"track_event,omitempty"`
// 		} `json:"msys"`
// 	}

// 	if err := json.Unmarshal(data, &payload); err != nil {
// 		return err
// 	}

// 	if len(payload) == 0 {
// 		return fmt.Errorf("empty payload")
// 	}

// 	var event *CommonEventFields
// 	if payload[0].Msys.MessageEvent != nil {
// 		event = payload[0].Msys.MessageEvent
// 	} else if payload[0].Msys.TrackEvent != nil {
// 		event = payload[0].Msys.TrackEvent
// 	} else {
// 		return fmt.Errorf("no recognized event type in payload")
// 	}

// 	*c = *event // Copy the unmarshaled data to the receiver

// 	fmt.Printf("Unmarshalled data: %+v\n", c)

// 	return c.saveToDatabase(data, headers)
// }

// func (c *CommonEventFields) saveToDatabase(eventData []byte, headers SparkPostWebhookHeaders) error {
// 	database.InitDB()
// 	db := database.GetDB()

// 	stmt, err := db.Prepare(`
// 		INSERT INTO sparkpost_events (
// 			event_type, message_id, transmission_id, event_data,
// 			accept_encoding, content_length, content_type, user_agent, x_forwarded_for,
// 			x_forwarded_host, x_forwarded_proto, x_sparkpost_signature, auth_header,
// 			timestamp, rcpt_to, ip_address, event_id, timestamp
// 		) VALUES (
// 			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
// 		)
// 	`)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	timeAsInt, _ := strconv.Atoi(c.Timestamp)

// 	_, err = stmt.Exec(
// 		c.Type,
// 		c.MessageID,
// 		c.TransmissionID,
// 		string(eventData),
// 		pq.Array(headers.AcceptEncoding),
// 		pq.Array(headers.ContentLength),
// 		pq.Array(headers.ContentType),
// 		pq.Array(headers.UserAgent),
// 		pq.Array(headers.XForwardedFor),
// 		pq.Array(headers.XForwardedHost),
// 		pq.Array(headers.XForwardedProto),
// 		pq.Array(headers.XSparkpostSignature),
// 		pq.Array(headers.Authorization),
// 		timeAsInt,
// 		c.RcptTo,
// 		c.IPAddress,
// 		c.EventID,
// 	)

// 	return err
// }

// func (e *MessageEvent) UnmarshalSparkPostEvent(data []byte, headers SparkPostWebhookHeaders) error {
// 	return e.CommonEventFields.UnmarshalSparkPostEvent(data, headers)
// }

// func (e *TrackEvent) UnmarshalSparkPostEvent(data []byte, headers SparkPostWebhookHeaders) error {
// 	return e.CommonEventFields.UnmarshalSparkPostEvent(data, headers)
// }

// type SparkPostPayload []struct {
// 	Msys struct {
// 		MessageEvent *MessageEvent `json:"message_event,omitempty"`
// 		TrackEvent   *TrackEvent   `json:"track_event,omitempty"`
// 	} `json:"msys"`
// }

// func ProcessSparkPostEvents(msg *sarama.ConsumerMessage) {
// 	var payload struct {
// 		Headers SparkPostWebhookHeaders `json:"headers"`
// 		Body    json.RawMessage         `json:"body"`
// 	}
// 	err := json.Unmarshal(msg.Value, &payload)
// 	if err != nil {
// 		fmt.Printf("Failed to unmarshal message: %v\n", err)
// 		return
// 	}
// 	var sparkPostPayload SparkPostPayload

// 	err = json.Unmarshal(payload.Body, &sparkPostPayload)
// 	if err != nil {
// 		fmt.Printf("Error unmarshaling JSON: %v\n", err)
// 		return
// 	}

// 	for _, event := range sparkPostPayload {
// 		var sparkPostEvent SparkPostEventUnmarshaler

// 		if event.Msys.MessageEvent != nil {
// 			sparkPostEvent = event.Msys.MessageEvent
// 			// fmt.Printf("Message Event Type: %s\n", event.Msys.MessageEvent.Type)
// 		} else if event.Msys.TrackEvent != nil {
// 			sparkPostEvent = event.Msys.TrackEvent
// 			// fmt.Printf("Track Event Type: %s\n", event.Msys.TrackEvent.Type)
// 		} else {
// 			fmt.Println("Unknown event type")
// 			continue
// 		}

// 		err = sparkPostEvent.UnmarshalSparkPostEvent(payload.Body, payload.Headers)
// 		if err != nil {
// 			fmt.Printf("Error unmarshaling and saving event to database: %v\n", err)
// 		}
// 	}
// }
