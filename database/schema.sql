--
-- PostgreSQL database dump
--

-- Dumped from database version 14.13 (Debian 14.13-1.pgdg120+1)
-- Dumped by pg_dump version 14.13 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: update_event_counts(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_event_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Set the unix_timestamp
        NEW.unix_timestamp := EXTRACT(EPOCH FROM NEW.created_at);
        
        -- Update only the latest event record
        WITH latest_event AS (
            SELECT id
            FROM events
            WHERE provider = NEW.provider 
              AND event_type = NEW.event_type 
              AND message_id = NEW.message_id
              AND DATE(timestamp) = DATE(NEW.timestamp)
            ORDER BY timestamp DESC
            LIMIT 1
        )
        UPDATE events e
        SET total_count = total_count + 1,
            daily_count = daily_count + 1,
            hourly_count = CASE WHEN DATE_TRUNC('hour', e.timestamp) = DATE_TRUNC('hour', NEW.timestamp) THEN hourly_count + 1 ELSE 1 END,
            unique_recipient_count = unique_recipient_count + (CASE WHEN e.recipient_email = NEW.recipient_email THEN 0 ELSE 1 END),
            unique_sender_count = unique_sender_count + (CASE WHEN e.sender_email = NEW.sender_email THEN 0 ELSE 1 END),
            campaign_count = campaign_count + (CASE WHEN e.campaign_id = NEW.campaign_id THEN 0 ELSE 1 END),
            domain_count = domain_count + (CASE WHEN e.recipient_domain = NEW.recipient_domain THEN 0 ELSE 1 END),
            updated_at = CURRENT_TIMESTAMP,
            unix_timestamp = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)
        FROM latest_event
        WHERE e.id = latest_event.id;
        
        IF FOUND THEN
            RETURN NULL;
        END IF;
    END IF;
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: email_service_providers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_service_providers (
    esp_id integer NOT NULL,
    user_id integer NOT NULL,
    provider_name character varying(100) NOT NULL,
    api_key character varying(255) NOT NULL,
    sending_domains text[] NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    sendgrid_verification_key character varying(180),
    sparkpost_webhook_user character varying(100),
    sparkpost_webhook_password character varying(100),
    socketlabs_secret_key character varying(100),
    postmark_webhook_user character varying(100),
    postmark_webhook_password character varying(100)
);


--
-- Name: email_service_providers_esp_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.email_service_providers_esp_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: email_service_providers_esp_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.email_service_providers_esp_id_seq OWNED BY public.email_service_providers.esp_id;


--
-- Name: postmark_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.postmark_events (
    id integer NOT NULL,
    record_type character varying(255) NOT NULL,
    server_id integer,
    message_id character varying(255),
    recipient character varying(255),
    tag character varying(255),
    delivered_at timestamp without time zone,
    details text,
    metadata jsonb,
    provider character varying(255),
    event character varying(255),
    event_data jsonb,
    accept_encoding text[],
    content_length text[],
    content_type text[],
    expect text[],
    user_agent text[],
    x_forwarded_for text[],
    x_forwarded_host text[],
    x_forwarded_proto text[],
    x_pm_retries_remaining text[],
    x_pm_webhook_event_id text[],
    x_pm_webhook_trace_id text[],
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    auth_header character varying(255) NOT NULL
);


--
-- Name: postmarkeventwithheaders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.postmarkeventwithheaders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: postmarkeventwithheaders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.postmarkeventwithheaders_id_seq OWNED BY public.postmark_events.id;


--
-- Name: sendgrid_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sendgrid_events (
    id integer NOT NULL,
    provider character varying(255),
    email character varying(255),
    "timestamp" bigint,
    smtp_id character varying(255),
    event character varying(255),
    category text[],
    sg_event_id character varying(255),
    sg_message_id character varying(255),
    accept_encoding text[],
    content_length text[],
    content_type text[],
    user_agent text[],
    x_forwarded_for text[],
    x_forwarded_host text[],
    x_forwarded_proto text[],
    x_twilio_email_event_webhook_signature text[],
    x_twilio_email_event_webhook_timestamp text[]
);


--
-- Name: sendgrideventwithheaders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sendgrideventwithheaders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sendgrideventwithheaders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sendgrideventwithheaders_id_seq OWNED BY public.sendgrid_events.id;


--
-- Name: socketlabs_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.socketlabs_events (
    id integer NOT NULL,
    event_type character varying(255) NOT NULL,
    date_time timestamp without time zone NOT NULL,
    mailing_id character varying(255),
    message_id character varying(255),
    address character varying(255),
    server_id integer,
    subaccount_id integer,
    ip_pool_id integer,
    secret_key character varying(255),
    event_data jsonb,
    accept_encoding text[],
    content_length text[],
    content_type text[],
    user_agent text[],
    x_forwarded_for text[],
    x_forwarded_host text[],
    x_forwarded_proto text[],
    x_socketlabs_signature text[],
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: socketlabs_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.socketlabs_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: socketlabs_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.socketlabs_events_id_seq OWNED BY public.socketlabs_events.id;


--
-- Name: sparkpost_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sparkpost_events (
    id integer NOT NULL,
    event_type character varying(255) NOT NULL,
    message_id character varying(255),
    transmission_id character varying(255),
    event_data jsonb,
    accept_encoding text[],
    content_length text[],
    content_type text[],
    user_agent text[],
    x_forwarded_for text[],
    x_forwarded_host text[],
    x_forwarded_proto text[],
    x_sparkpost_signature text[],
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "timestamp" bigint,
    rcpt_to character varying(255),
    ip_address character varying(45),
    event_id character varying(255),
    auth_header character varying(255) NOT NULL
);


--
-- Name: sparkpost_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sparkpost_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sparkpost_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sparkpost_events_id_seq OWNED BY public.sparkpost_events.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    user_id integer NOT NULL,
    username character varying(50) NOT NULL,
    email character varying(255) NOT NULL,
    first_name character varying(50),
    last_name character varying(50),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.user_id;


--
-- Name: email_service_providers esp_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers ALTER COLUMN esp_id SET DEFAULT nextval('public.email_service_providers_esp_id_seq'::regclass);


--
-- Name: postmark_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.postmark_events ALTER COLUMN id SET DEFAULT nextval('public.postmarkeventwithheaders_id_seq'::regclass);


--
-- Name: sendgrid_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sendgrid_events ALTER COLUMN id SET DEFAULT nextval('public.sendgrideventwithheaders_id_seq'::regclass);


--
-- Name: socketlabs_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.socketlabs_events ALTER COLUMN id SET DEFAULT nextval('public.socketlabs_events_id_seq'::regclass);


--
-- Name: sparkpost_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sparkpost_events ALTER COLUMN id SET DEFAULT nextval('public.sparkpost_events_id_seq'::regclass);


--
-- Name: users user_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN user_id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- Name: email_service_providers email_service_providers_api_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT email_service_providers_api_key_key UNIQUE (api_key);


--
-- Name: email_service_providers email_service_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT email_service_providers_pkey PRIMARY KEY (esp_id);


--
-- Name: postmark_events postmarkeventwithheaders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.postmark_events
    ADD CONSTRAINT postmarkeventwithheaders_pkey PRIMARY KEY (id);


--
-- Name: sendgrid_events sendgrideventwithheaders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sendgrid_events
    ADD CONSTRAINT sendgrideventwithheaders_pkey PRIMARY KEY (id);


--
-- Name: socketlabs_events socketlabs_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.socketlabs_events
    ADD CONSTRAINT socketlabs_events_pkey PRIMARY KEY (id);


--
-- Name: sparkpost_events sparkpost_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sparkpost_events
    ADD CONSTRAINT sparkpost_events_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_sendgrid_events_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sendgrid_events_email ON public.sendgrid_events USING btree (email);


--
-- Name: idx_sendgrid_events_email_timestamp; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sendgrid_events_email_timestamp ON public.sendgrid_events USING btree (email, "timestamp");


--
-- Name: idx_sendgrid_events_event; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sendgrid_events_event ON public.sendgrid_events USING btree (event);


--
-- Name: idx_sendgrid_events_sg_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sendgrid_events_sg_event_id ON public.sendgrid_events USING btree (sg_event_id);


--
-- Name: email_service_providers email_service_providers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT email_service_providers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

