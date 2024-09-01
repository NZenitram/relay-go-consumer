CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    total_count BIGINT DEFAULT 1,
    daily_count BIGINT DEFAULT 1,
    hourly_count BIGINT DEFAULT 1,
    unique_recipient_count BIGINT DEFAULT 1,
    unique_sender_count BIGINT DEFAULT 1,
    campaign_count BIGINT DEFAULT 1,
    domain_count BIGINT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    message_id VARCHAR(255),
    recipient_email VARCHAR(255),
    sender_email VARCHAR(255),
    subject TEXT,
    timestamp TIMESTAMP WITH TIME ZONE,
    campaign_id VARCHAR(255),
    recipient_domain VARCHAR(255),
    unix_timestamp BIGINT
);

-- Create indices for faster queries
CREATE INDEX idx_provider_event_type ON events (provider, event_type);
CREATE INDEX idx_timestamp ON events (timestamp);
CREATE INDEX idx_campaign_id ON events (campaign_id);
CREATE INDEX idx_recipient_domain ON events (recipient_domain);
CREATE INDEX idx_unix_timestamp ON events (unix_timestamp);

-- Create a unique constraint
CREATE UNIQUE INDEX idx_unique_event ON events (provider, event_type, message_id, DATE(timestamp));

-- Function to update counts and set unix_timestamp
CREATE OR REPLACE FUNCTION update_event_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Set the unix_timestamp
        NEW.unix_timestamp := EXTRACT(EPOCH FROM NEW.created_at);
        
        -- Check if a record already exists for this event on this day
        UPDATE events
        SET total_count = total_count + 1,
            daily_count = daily_count + 1,
            hourly_count = CASE WHEN DATE_TRUNC('hour', timestamp) = DATE_TRUNC('hour', NEW.timestamp) THEN hourly_count + 1 ELSE 1 END,
            unique_recipient_count = unique_recipient_count + (CASE WHEN recipient_email = NEW.recipient_email THEN 0 ELSE 1 END),
            unique_sender_count = unique_sender_count + (CASE WHEN sender_email = NEW.sender_email THEN 0 ELSE 1 END),
            campaign_count = campaign_count + (CASE WHEN campaign_id = NEW.campaign_id THEN 0 ELSE 1 END),
            domain_count = domain_count + (CASE WHEN recipient_domain = NEW.recipient_domain THEN 0 ELSE 1 END),
            updated_at = CURRENT_TIMESTAMP,
            unix_timestamp = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)
        WHERE provider = NEW.provider 
          AND event_type = NEW.event_type 
          AND message_id = NEW.message_id
          AND DATE(timestamp) = DATE(NEW.timestamp);
        
        IF FOUND THEN
            RETURN NULL;
        END IF;
    END IF;
    RETURN NEW;
END;
$$
 LANGUAGE plpgsql;

-- Trigger to call the function before insert
CREATE TRIGGER before_insert_event
BEFORE INSERT ON events
FOR EACH ROW
EXECUTE FUNCTION update_event_counts();
