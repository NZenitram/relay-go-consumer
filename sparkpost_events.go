package main

import (
	"encoding/json"
	"fmt"
	"relay-go-consumer/database"

	"github.com/IBM/sarama"
	"github.com/lib/pq"
)

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
	ContentLength       []string `json:"Content-Length"`
	ContentType         []string `json:"Content-Type"`
	UserAgent           []string `json:"User-Agent"`
	XForwardedFor       []string `json:"X-Forwarded-For"`
	XForwardedHost      []string `json:"X-Forwarded-Host"`
	XForwardedProto     []string `json:"X-Forwarded-Proto"`
	XSparkpostSignature []string `json:"X-Sparkpost-Signature"`
}

// ... (keep all your existing struct definitions for GeoIP, UserAgentParsed, etc.)

type SparkPostEventUnmarshaler interface {
	UnmarshalSparkPostEvent(data []byte, headers SparkPostWebhookHeaders) error
}

func (c *CommonEventFields) UnmarshalSparkPostEvent(data []byte, headers SparkPostWebhookHeaders) error {
	fmt.Printf("Raw data: %s\n", string(data))

	var payload []struct {
		Msys struct {
			MessageEvent *CommonEventFields `json:"message_event,omitempty"`
			TrackEvent   *CommonEventFields `json:"track_event,omitempty"`
		} `json:"msys"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if len(payload) == 0 {
		return fmt.Errorf("empty payload")
	}

	var event *CommonEventFields
	if payload[0].Msys.MessageEvent != nil {
		event = payload[0].Msys.MessageEvent
	} else if payload[0].Msys.TrackEvent != nil {
		event = payload[0].Msys.TrackEvent
	} else {
		return fmt.Errorf("no recognized event type in payload")
	}

	*c = *event // Copy the unmarshaled data to the receiver

	fmt.Printf("Unmarshalled data: %+v\n", c)

	return c.saveToDatabase(data, headers)
}

func (c *CommonEventFields) saveToDatabase(eventData []byte, headers SparkPostWebhookHeaders) error {
	database.InitDB()
	db := database.GetDB()

	stmt, err := db.Prepare(`
		INSERT INTO sparkpost_events (
			event_type, message_id, transmission_id, event_data,
			accept_encoding, content_length, content_type, user_agent, x_forwarded_for,
			x_forwarded_host, x_forwarded_proto, x_sparkpost_signature,
			timestamp, rcpt_to, ip_address, event_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		c.Type,
		c.MessageID,
		c.TransmissionID,
		string(eventData),
		pq.Array(headers.AcceptEncoding),
		pq.Array(headers.ContentLength),
		pq.Array(headers.ContentType),
		pq.Array(headers.UserAgent),
		pq.Array(headers.XForwardedFor),
		pq.Array(headers.XForwardedHost),
		pq.Array(headers.XForwardedProto),
		pq.Array(headers.XSparkpostSignature),
		c.Timestamp,
		c.RcptTo,
		c.IPAddress,
		c.EventID,
	)

	return err
}

func (e *MessageEvent) UnmarshalSparkPostEvent(data []byte, headers SparkPostWebhookHeaders) error {
	return e.CommonEventFields.UnmarshalSparkPostEvent(data, headers)
}

func (e *TrackEvent) UnmarshalSparkPostEvent(data []byte, headers SparkPostWebhookHeaders) error {
	return e.CommonEventFields.UnmarshalSparkPostEvent(data, headers)
}

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
		var sparkPostEvent SparkPostEventUnmarshaler

		if event.Msys.MessageEvent != nil {
			sparkPostEvent = event.Msys.MessageEvent
			// fmt.Printf("Message Event Type: %s\n", event.Msys.MessageEvent.Type)
		} else if event.Msys.TrackEvent != nil {
			sparkPostEvent = event.Msys.TrackEvent
			// fmt.Printf("Track Event Type: %s\n", event.Msys.TrackEvent.Type)
		} else {
			fmt.Println("Unknown event type")
			continue
		}

		err = sparkPostEvent.UnmarshalSparkPostEvent(payload.Body, payload.Headers)
		if err != nil {
			fmt.Printf("Error unmarshaling and saving event to database: %v\n", err)
		}
	}
}

// // SparkPostEventUnmarshaler interface
// type SparkPostEventUnmarshaler interface {
// 	UnmarshalSparkPostEvent(data []byte) error
// }

// func (e *SparkPostEvent) UnmarshalJSON(data []byte) error {
// 	var wrapper struct {
// 		Msys struct {
// 			TrackEvent json.RawMessage `json:"track_event"`
// 		} `json:"msys"`
// 	}

// 	if err := json.Unmarshal(data, &wrapper); err != nil {
// 		return err
// 	}

// 	return json.Unmarshal(wrapper.Msys.TrackEvent, e)
// }

// func ProcessSparkPostEvents(msg *sarama.ConsumerMessage) {
// 	var payload struct {
// 		Headers json.RawMessage `json:"headers"`
// 		Body    []struct {
// 			Msys struct {
// 				MessageEvent json.RawMessage `json:"message_event"`
// 			} `json:"msys"`
// 		} `json:"body"`
// 	}

// 	err := json.Unmarshal(msg.Value, &payload)
// 	if err != nil {
// 		log.Printf("Failed to unmarshal SparkPost event wrapper: %v", err)
// 		return
// 	}

// 	if len(payload.Body) == 0 {
// 		log.Printf("No events in payload")
// 		return
// 	}

// 	eventData := payload.Body[0].Msys.MessageEvent

// 	var baseEvent SparkPostEvent
// 	err = json.Unmarshal(eventData, &baseEvent)
// 	if err != nil {
// 		log.Printf("Failed to unmarshal base event: %v", err)
// 		log.Printf("Raw event data: %s", string(eventData))
// 		return
// 	}

// 	baseEvent.Provider = "SparkPost"

// 	var event SparkPostEventUnmarshaler

// 	switch baseEvent.Type {
// 	case "bounce":
// 		event = &SparkPostBounceEvent{SparkPostEvent: baseEvent}
// 	case "delivery":
// 		event = &SparkPostDeliveryEvent{SparkPostEvent: baseEvent}
// 	case "injection":
// 		event = &SparkPostInjectionEvent{SparkPostEvent: baseEvent}
// 	case "spam_complaint":
// 		event = &SparkPostSpamComplaintEvent{SparkPostEvent: baseEvent}
// 	case "out_of_band":
// 		event = &SparkPostOutOfBandEvent{SparkPostEvent: baseEvent}
// 	case "policy_rejection":
// 		event = &SparkPostPolicyRejectionEvent{SparkPostEvent: baseEvent}
// 	case "delay":
// 		event = &SparkPostDelayEvent{SparkPostEvent: baseEvent}
// 	case "click":
// 		event = &SparkPostClickEvent{SparkPostEvent: baseEvent}
// 	case "open":
// 		event = &SparkPostOpenEvent{SparkPostEvent: baseEvent}
// 	case "initial_open":
// 		event = &SparkPostInitialOpenEvent{SparkPostEvent: baseEvent}
// 	case "amp_click":
// 		event = &SparkPostAmpClickEvent{SparkPostEvent: baseEvent}
// 	case "amp_open":
// 		event = &SparkPostAmpOpenEvent{SparkPostEvent: baseEvent}
// 	case "amp_initial_open":
// 		event = &SparkPostAmpInitialOpenEvent{SparkPostEvent: baseEvent}
// 	case "generation_failure":
// 		event = &SparkPostGenerationFailureEvent{SparkPostEvent: baseEvent}
// 	case "generation_rejection":
// 		event = &SparkPostGenerationRejectionEvent{SparkPostEvent: baseEvent}
// 	case "unsubscribe":
// 		event = &SparkPostUnsubscribeEvent{SparkPostEvent: baseEvent}
// 	case "link_unsubscribe":
// 		event = &SparkPostLinkUnsubscribeEvent{SparkPostUnsubscribeEvent: SparkPostUnsubscribeEvent{SparkPostEvent: baseEvent}}
// 	case "relay_injection":
// 		event = &SparkPostRelayInjectionEvent{SparkPostEvent: baseEvent}
// 	case "relay_rejection":
// 		event = &SparkPostRelayRejectionEvent{SparkPostEvent: baseEvent}
// 	case "relay_delivery":
// 		event = &SparkPostRelayDeliveryEvent{SparkPostEvent: baseEvent}
// 	case "relay_tempfail":
// 		event = &SparkPostRelayTempFailEvent{SparkPostEvent: baseEvent}
// 	case "relay_permfail":
// 		event = &SparkPostRelayPermFailEvent{SparkPostEvent: baseEvent}
// 	case "ab_test_completed":
// 		event = &SparkPostABTestCompletedEvent{SparkPostEvent: baseEvent}
// 	case "ab_test_cancelled":
// 		event = &SparkPostABTestCancelledEvent{SparkPostEvent: baseEvent}
// 	case "ingest_success":
// 		event = &SparkPostIngestSuccessEvent{SparkPostEvent: baseEvent}
// 	case "ingest_error":
// 		event = &SparkPostIngestErrorEvent{SparkPostIngestSuccessEvent: SparkPostIngestSuccessEvent{SparkPostEvent: baseEvent}}
// 	default:
// 		log.Printf("Unknown event type: %s", baseEvent.Type)
// 		return
// 	}

// 	err = event.UnmarshalSparkPostEvent(eventData)
// 	if err != nil {
// 		log.Printf("Failed to unmarshal event: %v", err)
// 		log.Printf("Raw event data: %s", string(eventData))
// 		return
// 	}

// 	fmt.Printf("Event: %+v\n", event)
// 	saveToDatabase(event)
// }

// // Implement UnmarshalSparkPostEvent for each event type
// func (e *SparkPostBounceEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// // Repeat for all other event types...

// func (e *SparkPostIngestErrorEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostDeliveryEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostInjectionEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostSpamComplaintEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostOutOfBandEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostPolicyRejectionEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostDelayEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostClickEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostOpenEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostInitialOpenEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostAmpClickEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostAmpOpenEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostAmpInitialOpenEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostGenerationFailureEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostGenerationRejectionEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostUnsubscribeEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostLinkUnsubscribeEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostRelayInjectionEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostRelayRejectionEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostRelayDeliveryEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostRelayTempFailEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostRelayPermFailEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostABTestCompletedEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostABTestCancelledEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// func (e *SparkPostIngestSuccessEvent) UnmarshalSparkPostEvent(data []byte) error {
// 	return json.Unmarshal(data, e)
// }

// // Base struct for common fields
// type SparkPostEvent struct {
// 	Provider              string                 `json:"provider,omitempty"`
// 	AmpEnabled            bool                   `json:"amp_enabled,omitempty"`
// 	BounceClass           string                 `json:"bounce_class,omitempty"`
// 	CampaignID            string                 `json:"campaign_id,omitempty"`
// 	ClickTracking         bool                   `json:"click_tracking,omitempty"`
// 	CustomerID            string                 `json:"customer_id,omitempty"`
// 	DelvMethod            string                 `json:"delv_method,omitempty"`
// 	DeviceToken           string                 `json:"device_token,omitempty"`
// 	ErrorCode             string                 `json:"error_code,omitempty"`
// 	EventID               string                 `json:"event_id,omitempty"`
// 	FriendlyFrom          string                 `json:"friendly_from,omitempty"`
// 	InitialPixel          bool                   `json:"initial_pixel,omitempty"`
// 	InjectionTime         string                 `json:"injection_time,omitempty"`
// 	IPAddress             string                 `json:"ip_address,omitempty"`
// 	IPPool                string                 `json:"ip_pool,omitempty"`
// 	MailboxProvider       string                 `json:"mailbox_provider,omitempty"`
// 	MailboxProviderRegion string                 `json:"mailbox_provider_region,omitempty"`
// 	MessageID             string                 `json:"message_id,omitempty"`
// 	MsgFrom               string                 `json:"msg_from,omitempty"`
// 	MsgSize               string                 `json:"msg_size,omitempty"`
// 	NumRetries            string                 `json:"num_retries,omitempty"`
// 	OpenTracking          bool                   `json:"open_tracking,omitempty"`
// 	RcptMeta              map[string]interface{} `json:"rcpt_meta,omitempty"`
// 	RcptTags              []string               `json:"rcpt_tags,omitempty"`
// 	RcptTo                string                 `json:"rcpt_to,omitempty"`
// 	RcptHash              string                 `json:"rcpt_hash,omitempty"`
// 	RawRcptTo             string                 `json:"raw_rcpt_to,omitempty"`
// 	RcptType              string                 `json:"rcpt_type,omitempty"`
// 	RawReason             string                 `json:"raw_reason,omitempty"`
// 	Reason                string                 `json:"reason,omitempty"`
// 	RecipientDomain       string                 `json:"recipient_domain"`
// 	RecvMethod            string                 `json:"recv_method,omitempty"`
// 	RoutingDomain         string                 `json:"routing_domain"`
// 	ScheduledTime         string                 `json:"scheduled_time,omitempty"`
// 	SendingIP             string                 `json:"sending_ip,omitempty"`
// 	SmsCoding             string                 `json:"sms_coding,omitempty"`
// 	SmsDst                string                 `json:"sms_dst,omitempty"`
// 	SmsDstNpi             string                 `json:"sms_dst_npi,omitempty"`
// 	SmsDstTon             string                 `json:"sms_dst_ton,omitempty"`
// 	SmsSrc                string                 `json:"sms_src,omitempty"`
// 	SmsSrcNpi             string                 `json:"sms_src_npi,omitempty"`
// 	SmsSrcTon             string                 `json:"sms_src_ton,omitempty"`
// 	SubaccountID          string                 `json:"subaccount_id,omitempty"`
// 	Subject               string                 `json:"subject,omitempty"`
// 	TemplateID            string                 `json:"template_id,omitempty"`
// 	TemplateVersion       string                 `json:"template_version,omitempty"`
// 	Timestamp             string                 `json:"timestamp,omitempty"`
// 	Transactional         string                 `json:"transactional,omitempty"`
// 	TransmissionID        string                 `json:"transmission_id"`
// 	Type                  string                 `json:"type,omitempty"`
// 	ABTestID              string                 `json:"ab_test_id,omitempty"`
// 	ABTestVersion         string                 `json:"ab_test_version,omitempty"`
// 	OutboundTLS           string                 `json:"outbound_tls,omitempty"`
// 	QueueTime             string                 `json:"queue_time,omitempty"`
// 	SMSRemoteIDs          []string               `json:"sms_remoteids,omitempty"`
// 	SMSSegments           int                    `json:"sms_segments,omitempty"`
// 	GeoIP                 *GeoIP                 `json:"geo_ip,omitempty"`
// 	UserAgentParsed       *UserAgentParsed       `json:"user_agent_parsed,omitempty"`
// }

// type SparkPostBounceEvent struct {
// 	SparkPostEvent
// }

// type SparkPostDeliveryEvent struct {
// 	SparkPostEvent
// 	OutboundTLS  string   `json:"outbound_tls,omitempty"`
// 	QueueTime    string   `json:"queue_time,omitempty"`
// 	SmsRemoteIDs []string `json:"sms_remoteids,omitempty"`
// 	SmsSegments  int      `json:"sms_segments,omitempty"`
// }

// type SparkPostInjectionEvent struct {
// 	SparkPostEvent
// 	SmsText string `json:"sms_text,omitempty"`
// }

// type SparkPostSpamComplaintEvent struct {
// 	SparkPostEvent
// 	FbType   string `json:"fbtype,omitempty"`
// 	ReportBy string `json:"report_by,omitempty"`
// 	ReportTo string `json:"report_to,omitempty"`
// }

// type SparkPostOutOfBandEvent struct {
// 	SparkPostEvent
// }

// type SparkPostPolicyRejectionEvent struct {
// 	SparkPostEvent
// }

// type SparkPostDelayEvent struct {
// 	SparkPostEvent
// 	SmsRemoteIDs []string `json:"sms_remoteids,omitempty"`
// 	SmsSegments  int      `json:"sms_segments,omitempty"`
// }

// type GeoIP struct {
// 	IP             string `json:"IP,omitempty"`
// 	City           string `json:"City,omitempty"`
// 	Country        string `json:"Country,omitempty"`
// 	CountryISOCode string `json:"CountryISOCode,omitempty"`
// 	Region         string `json:"Region,omitempty"`
// 	RegionISOCode  string `json:"RegionISOCode,omitempty"`
// 	Zip            string `json:"Zip,omitempty"`
// 	Coords         string `json:"Coords,omitempty"`
// }

// type UserAgentParsed struct {
// 	AgentFamily  string `json:"agent_family,omitempty"`
// 	DeviceBrand  string `json:"device_brand,omitempty"`
// 	DeviceFamily string `json:"device_family,omitempty"`
// 	OSFamily     string `json:"os_family,omitempty"`
// 	OSVersion    string `json:"os_version,omitempty"`
// 	IsMobile     bool   `json:"is_mobile,omitempty"`
// 	IsProxy      bool   `json:"is_proxy,omitempty"`
// 	IsPrefetched bool   `json:"is_prefetched,omitempty"`
// }

// type SparkPostClickEvent struct {
// 	SparkPostEvent
// 	TargetLinkName string `json:"target_link_name,omitempty"`
// 	TargetLinkURL  string `json:"target_link_url,omitempty"`
// }

// type SparkPostOpenEvent struct {
// 	SparkPostEvent
// }

// type SparkPostInitialOpenEvent struct {
// 	SparkPostEvent
// }

// type SparkPostAmpClickEvent struct {
// 	SparkPostEvent
// 	TargetLinkName string `json:"target_link_name,omitempty"`
// 	TargetLinkURL  string `json:"target_link_url,omitempty"`
// }

// type SparkPostAmpOpenEvent struct {
// 	SparkPostEvent
// }

// type SparkPostAmpInitialOpenEvent struct {
// 	SparkPostEvent
// }

// type SparkPostGenerationFailureEvent struct {
// 	SparkPostEvent
// 	RcptSubs map[string]string `json:"rcpt_subs,omitempty"`
// }

// type SparkPostGenerationRejectionEvent struct {
// 	SparkPostEvent
// 	BounceClass string `json:"bounce_class,omitempty"`
// }

// type SparkPostUnsubscribeEvent struct {
// 	SparkPostEvent
// 	MailFrom string `json:"mailfrom,omitempty"`
// }

// type SparkPostLinkUnsubscribeEvent struct {
// 	SparkPostUnsubscribeEvent
// 	TargetLinkName string `json:"target_link_name,omitempty"`
// 	TargetLinkURL  string `json:"target_link_url,omitempty"`
// 	UserAgent      string `json:"user_agent,omitempty"`
// }

// type SparkPostRelayInjectionEvent struct {
// 	SparkPostEvent
// 	RelayID     string `json:"relay_id,omitempty"`
// 	Origination string `json:"origination,omitempty"`
// }

// type SparkPostRelayRejectionEvent struct {
// 	SparkPostEvent
// 	RelayID     string `json:"relay_id,omitempty"`
// 	RawReason   string `json:"raw_reason,omitempty"`
// 	Reason      string `json:"reason,omitempty"`
// 	BounceClass string `json:"bounce_class,omitempty"`
// }

// type SparkPostRelayDeliveryEvent struct {
// 	SparkPostEvent
// 	RelayID    string `json:"relay_id,omitempty"`
// 	QueueTime  string `json:"queue_time,omitempty"`
// 	NumRetries string `json:"num_retries,omitempty"`
// 	DelvMethod string `json:"delv_method,omitempty"`
// }

// type SparkPostRelayTempFailEvent struct {
// 	SparkPostEvent
// 	RelayID   string `json:"relay_id,omitempty"`
// 	RawReason string `json:"raw_reason,omitempty"`
// 	Reason    string `json:"reason,omitempty"`
// 	ErrorCode string `json:"error_code,omitempty"`
// }

// type SparkPostRelayPermFailEvent struct {
// 	SparkPostEvent
// 	RelayID     string `json:"relay_id,omitempty"`
// 	RawReason   string `json:"raw_reason,omitempty"`
// 	Reason      string `json:"reason,omitempty"`
// 	ErrorCode   string `json:"error_code,omitempty"`
// 	OutboundTLS string `json:"outbound_tls,omitempty"`
// }

// type SparkPostABTestCompletedEvent struct {
// 	SparkPostEvent
// 	ABTest ABTestDetails `json:"ab_test"`
// }

// type SparkPostABTestCancelledEvent struct {
// 	SparkPostEvent
// 	ABTest ABTestDetails `json:"ab_test"`
// }

// type SparkPostIngestSuccessEvent struct {
// 	SparkPostEvent
// 	BatchID             string `json:"batch_id"`
// 	ExpirationTimestamp string `json:"expiration_timestamp"`
// 	NumberSucceeded     int    `json:"number_succeeded"`
// 	NumberDuplicates    int    `json:"number_duplicates"`
// }

// type SparkPostIngestErrorEvent struct {
// 	SparkPostIngestSuccessEvent
// 	ErrorType    string `json:"error_type"`
// 	NumberFailed int    `json:"number_failed"`
// 	Retryable    bool   `json:"retryable"`
// 	Href         string `json:"href"`
// }

// type ABTestDetails struct {
// 	ID                string         `json:"id"`
// 	Name              string         `json:"name"`
// 	Version           int            `json:"version"`
// 	TestMode          string         `json:"test_mode"`
// 	EngagementMetric  string         `json:"engagement_metric"`
// 	DefaultTemplate   TemplateInfo   `json:"default_template"`
// 	Variants          []TemplateInfo `json:"variants"`
// 	WinningTemplateID string         `json:"winning_template_id,omitempty"`
// }

// type TemplateInfo struct {
// 	TemplateID         string  `json:"template_id"`
// 	CountUniqueClicked int     `json:"count_unique_clicked"`
// 	CountAccepted      int     `json:"count_accepted"`
// 	EngagementRate     float64 `json:"engagement_rate"`
// }
