package main

import (
	"encoding/base64"
	"log"

	sp "github.com/SparkPost/gosparkpost"
)

func SendEmailWithSparkPost(emailMessage EmailMessage) {
	// Get the API key from the credentials
	apiKey := emailMessage.Credentials.SparkpostAPIKey
	if apiKey == "" {
		log.Fatal("Missing SparkPost API key in credentials")
	}

	// Configure SparkPost client
	cfg := &sp.Config{
		BaseUrl:    "https://api.sparkpost.com",
		ApiKey:     apiKey,
		ApiVersion: 1,
	}
	var client sp.Client
	err := client.Init(cfg)
	if err != nil {
		log.Fatalf("SparkPost client init failed: %s\n", err)
	}

	// Prepare recipients
	recipients := make([]sp.Recipient, len(emailMessage.To))
	for i, addr := range emailMessage.To {
		recipients[i] = sp.Recipient{
			Address: sp.Address{
				Email: addr.Email,
			},
		}
	}

	// Prepare attachments
	attachments := make([]sp.Attachment, len(emailMessage.Attachments))
	for i, att := range emailMessage.Attachments {
		content, err := base64.StdEncoding.DecodeString(att.Content)
		if err != nil {
			log.Printf("Failed to decode attachment content: %v", err)
			continue
		}
		attachments[i] = sp.Attachment{
			Filename: att.Name,
			MIMEType: att.ContentType,
			B64Data:  string(content),
		}
	}

	// Create a Transmission
	tx := &sp.Transmission{
		Recipients: recipients,
		Content: sp.Content{
			From:        sp.Address{Email: emailMessage.From.Email, Name: emailMessage.From.Name},
			Subject:     emailMessage.Subject,
			HTML:        emailMessage.HtmlBody,
			Text:        emailMessage.TextBody,
			Headers:     emailMessage.Headers,
			Attachments: attachments,
		},
	}

	// Send the email
	id, _, err := client.Send(tx)
	if err != nil {
		log.Fatalf("Failed to send email with SparkPost: %v", err)
	}

	log.Printf("Email sent with SparkPost. Transmission ID: %s", id)
}

// Base struct for common fields
type SparkPostEvent struct {
	AmpEnabled            bool                   `json:"amp_enabled,omitempty"`
	BounceClass           string                 `json:"bounce_class,omitempty"`
	CampaignID            string                 `json:"campaign_id,omitempty"`
	ClickTracking         bool                   `json:"click_tracking,omitempty"`
	CustomerID            string                 `json:"customer_id,omitempty"`
	DelvMethod            string                 `json:"delv_method,omitempty"`
	DeviceToken           string                 `json:"device_token,omitempty"`
	ErrorCode             string                 `json:"error_code,omitempty"`
	EventID               string                 `json:"event_id"`
	FriendlyFrom          string                 `json:"friendly_from"`
	InitialPixel          bool                   `json:"initial_pixel,omitempty"`
	InjectionTime         string                 `json:"injection_time"`
	IPAddress             string                 `json:"ip_address,omitempty"`
	IPPool                string                 `json:"ip_pool,omitempty"`
	MailboxProvider       string                 `json:"mailbox_provider,omitempty"`
	MailboxProviderRegion string                 `json:"mailbox_provider_region,omitempty"`
	MessageID             string                 `json:"message_id"`
	MsgFrom               string                 `json:"msg_from"`
	MsgSize               string                 `json:"msg_size"`
	NumRetries            string                 `json:"num_retries,omitempty"`
	OpenTracking          bool                   `json:"open_tracking,omitempty"`
	RcptMeta              map[string]interface{} `json:"rcpt_meta,omitempty"`
	RcptTags              []string               `json:"rcpt_tags,omitempty"`
	RcptTo                string                 `json:"rcpt_to"`
	RcptHash              string                 `json:"rcpt_hash"`
	RawRcptTo             string                 `json:"raw_rcpt_to,omitempty"`
	RcptType              string                 `json:"rcpt_type,omitempty"`
	RawReason             string                 `json:"raw_reason,omitempty"`
	Reason                string                 `json:"reason,omitempty"`
	RecipientDomain       string                 `json:"recipient_domain"`
	RecvMethod            string                 `json:"recv_method,omitempty"`
	RoutingDomain         string                 `json:"routing_domain"`
	ScheduledTime         string                 `json:"scheduled_time,omitempty"`
	SendingIP             string                 `json:"sending_ip,omitempty"`
	SmsCoding             string                 `json:"sms_coding,omitempty"`
	SmsDst                string                 `json:"sms_dst,omitempty"`
	SmsDstNpi             string                 `json:"sms_dst_npi,omitempty"`
	SmsDstTon             string                 `json:"sms_dst_ton,omitempty"`
	SmsSrc                string                 `json:"sms_src,omitempty"`
	SmsSrcNpi             string                 `json:"sms_src_npi,omitempty"`
	SmsSrcTon             string                 `json:"sms_src_ton,omitempty"`
	SubaccountID          string                 `json:"subaccount_id,omitempty"`
	Subject               string                 `json:"subject"`
	TemplateID            string                 `json:"template_id,omitempty"`
	TemplateVersion       string                 `json:"template_version,omitempty"`
	Timestamp             string                 `json:"timestamp"`
	Transactional         string                 `json:"transactional,omitempty"`
	TransmissionID        string                 `json:"transmission_id"`
	Type                  string                 `json:"type"`
}

type SparkPostBounceEvent struct {
	SparkPostEvent
}

type SparkPostDeliveryEvent struct {
	SparkPostEvent
	OutboundTLS  string   `json:"outbound_tls,omitempty"`
	QueueTime    string   `json:"queue_time,omitempty"`
	SmsRemoteIDs []string `json:"sms_remoteids,omitempty"`
	SmsSegments  int      `json:"sms_segments,omitempty"`
}

type SparkPostInjectionEvent struct {
	SparkPostEvent
	SmsText string `json:"sms_text,omitempty"`
}

type SparkPostSpamComplaintEvent struct {
	SparkPostEvent
	FbType   string `json:"fbtype,omitempty"`
	ReportBy string `json:"report_by,omitempty"`
	ReportTo string `json:"report_to,omitempty"`
}

type SparkPostOutOfBandEvent struct {
	SparkPostEvent
}

type SparkPostPolicyRejectionEvent struct {
	SparkPostEvent
}

type SparkPostDelayEvent struct {
	SparkPostEvent
	SmsRemoteIDs []string `json:"sms_remoteids,omitempty"`
	SmsSegments  int      `json:"sms_segments,omitempty"`
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

type SparkPostClickEvent struct {
	SparkPostEvent
	TargetLinkName string `json:"target_link_name,omitempty"`
	TargetLinkURL  string `json:"target_link_url,omitempty"`
}

type SparkPostOpenEvent struct {
	SparkPostEvent
}

type SparkPostInitialOpenEvent struct {
	SparkPostEvent
}

type SparkPostAmpClickEvent struct {
	SparkPostEvent
	TargetLinkName string `json:"target_link_name,omitempty"`
	TargetLinkURL  string `json:"target_link_url,omitempty"`
}

type SparkPostAmpOpenEvent struct {
	SparkPostEvent
}

type SparkPostAmpInitialOpenEvent struct {
	SparkPostEvent
}

type SparkPostGenerationFailureEvent struct {
	SparkPostEvent
	RcptSubs map[string]string `json:"rcpt_subs,omitempty"`
}

type SparkPostGenerationRejectionEvent struct {
	SparkPostEvent
	BounceClass string `json:"bounce_class,omitempty"`
}

type SparkPostUnsubscribeEvent struct {
	SparkPostEvent
	MailFrom string `json:"mailfrom,omitempty"`
}

type SparkPostLinkUnsubscribeEvent struct {
	SparkPostUnsubscribeEvent
	TargetLinkName string `json:"target_link_name,omitempty"`
	TargetLinkURL  string `json:"target_link_url,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
}

type SparkPostRelayInjectionEvent struct {
	SparkPostEvent
	RelayID     string `json:"relay_id,omitempty"`
	Origination string `json:"origination,omitempty"`
}

type SparkPostRelayRejectionEvent struct {
	SparkPostEvent
	RelayID     string `json:"relay_id,omitempty"`
	RawReason   string `json:"raw_reason,omitempty"`
	Reason      string `json:"reason,omitempty"`
	BounceClass string `json:"bounce_class,omitempty"`
}

type SparkPostRelayDeliveryEvent struct {
	SparkPostEvent
	RelayID    string `json:"relay_id,omitempty"`
	QueueTime  string `json:"queue_time,omitempty"`
	NumRetries string `json:"num_retries,omitempty"`
	DelvMethod string `json:"delv_method,omitempty"`
}

type SparkPostRelayTempFailEvent struct {
	SparkPostEvent
	RelayID   string `json:"relay_id,omitempty"`
	RawReason string `json:"raw_reason,omitempty"`
	Reason    string `json:"reason,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

type SparkPostRelayPermFailEvent struct {
	SparkPostEvent
	RelayID     string `json:"relay_id,omitempty"`
	RawReason   string `json:"raw_reason,omitempty"`
	Reason      string `json:"reason,omitempty"`
	ErrorCode   string `json:"error_code,omitempty"`
	OutboundTLS string `json:"outbound_tls,omitempty"`
}

// SparkPostABTestCompletedEvent struct
type SparkPostABTestCompletedEvent struct {
	SparkPostEvent
	ABTest ABTestDetails `json:"ab_test"`
}

// SparkPostABTestCancelledEvent struct
type SparkPostABTestCancelledEvent struct {
	SparkPostEvent
	ABTest ABTestDetails `json:"ab_test"`
}

// SparkPostIngestSuccessEvent struct
type SparkPostIngestSuccessEvent struct {
	SparkPostEvent
	BatchID             string `json:"batch_id"`
	ExpirationTimestamp string `json:"expiration_timestamp"`
	NumberSucceeded     int    `json:"number_succeeded"`
	NumberDuplicates    int    `json:"number_duplicates"`
}

// SparkPostIngestErrorEvent struct
type SparkPostIngestErrorEvent struct {
	SparkPostIngestSuccessEvent
	ErrorType    string `json:"error_type"`
	NumberFailed int    `json:"number_failed"`
	Retryable    bool   `json:"retryable"`
	Href         string `json:"href"`
}

// Supporting struct for AB Test Details
type ABTestDetails struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Version           int            `json:"version"`
	TestMode          string         `json:"test_mode"`
	EngagementMetric  string         `json:"engagement_metric"`
	DefaultTemplate   TemplateInfo   `json:"default_template"`
	Variants          []TemplateInfo `json:"variants"`
	WinningTemplateID string         `json:"winning_template_id,omitempty"`
}

// Supporting struct for Template Information
type TemplateInfo struct {
	TemplateID         string  `json:"template_id"`
	CountUniqueClicked int     `json:"count_unique_clicked"`
	CountAccepted      int     `json:"count_accepted"`
	EngagementRate     float64 `json:"engagement_rate"`
}
