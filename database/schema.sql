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
    AS $_$
DECLARE
    count_column TEXT;
BEGIN
    -- Determine which count column to update based on the event type
    CASE NEW.event_type
        WHEN 'delivered' THEN count_column := 'delivered_count';
        WHEN 'opened' THEN count_column := 'opened_count';
        WHEN 'clicked' THEN count_column := 'clicked_count';
        WHEN 'bounced' THEN count_column := 'bounced_count';
        WHEN 'spam' THEN count_column := 'spam_count';
        WHEN 'unsubscribe' THEN count_column := 'unsubscribe_count';
        WHEN 'processed' THEN count_column := 'processed_count';
        WHEN 'dropped' THEN count_column := 'dropped_count';
        WHEN 'deferred' THEN count_column := 'deferred_count';
        ELSE RETURN NEW;  -- If it's not one of these events, don't update anything
    END CASE;

    -- Update the count for this user and event type
    EXECUTE format('
        WITH counts AS (
            SELECT %I AS count
            FROM %I
            WHERE user_id = $1 AND id < $2
            ORDER BY id DESC
            LIMIT 1
        )
        UPDATE %I
        SET %I = COALESCE((SELECT count FROM counts), 0) + 1
        WHERE id = $2',
        count_column, TG_TABLE_NAME, TG_TABLE_NAME, count_column)
    USING NEW.user_id, NEW.id;

    RETURN NEW;
END;
$_$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: email_service_providers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_service_providers (
    esp_id integer NOT NULL,
    user_id integer NOT NULL,
    provider_name character varying(100) NOT NULL,
    sending_domains text[] NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    sendgrid_verification_key character varying(180),
    sparkpost_webhook_user character varying(100),
    sparkpost_webhook_password character varying(100),
    socketlabs_secret_key character varying(100),
    postmark_webhook_user character varying(100),
    postmark_webhook_password character varying(100),
    socketlabs_server_id character varying(25)
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
    event_type character varying(255),
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
    auth_header character varying(255) NOT NULL,
    "timestamp" bigint,
    user_id integer,
    delivered_count integer DEFAULT 0,
    opened_count integer DEFAULT 0,
    clicked_count integer DEFAULT 0,
    bounced_count integer DEFAULT 0,
    spam_count integer DEFAULT 0,
    unsubscribe_count integer DEFAULT 0,
    processed_count integer DEFAULT 0,
    dropped_count integer DEFAULT 0,
    deferred_count integer DEFAULT 0
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
    event_type character varying(255),
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
    x_twilio_email_event_webhook_timestamp text[],
    user_id integer,
    delivered_count integer DEFAULT 0,
    opened_count integer DEFAULT 0,
    clicked_count integer DEFAULT 0,
    bounced_count integer DEFAULT 0,
    spam_count integer DEFAULT 0,
    unsubscribe_count integer DEFAULT 0,
    processed_count integer DEFAULT 0,
    dropped_count integer DEFAULT 0,
    deferred_count integer DEFAULT 0
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
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "timestamp" bigint,
    user_id integer,
    delivered_count integer DEFAULT 0,
    opened_count integer DEFAULT 0,
    clicked_count integer DEFAULT 0,
    bounced_count integer DEFAULT 0,
    spam_count integer DEFAULT 0,
    unsubscribe_count integer DEFAULT 0,
    processed_count integer DEFAULT 0,
    dropped_count integer DEFAULT 0,
    deferred_count integer DEFAULT 0
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
    auth_header character varying(255) NOT NULL,
    user_id integer,
    delivered_count integer DEFAULT 0,
    opened_count integer DEFAULT 0,
    clicked_count integer DEFAULT 0,
    bounced_count integer DEFAULT 0,
    spam_count integer DEFAULT 0,
    unsubscribe_count integer DEFAULT 0,
    processed_count integer DEFAULT 0,
    dropped_count integer DEFAULT 0,
    deferred_count integer DEFAULT 0
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
    id integer NOT NULL,
    username character varying(50) NOT NULL,
    email character varying(255) NOT NULL,
    first_name character varying(50),
    last_name character varying(50),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    api_key character varying(100)
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

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.id;


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
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


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
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_email_service_providers_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_service_providers_user_id ON public.email_service_providers USING btree (user_id);


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

CREATE INDEX idx_sendgrid_events_event ON public.sendgrid_events USING btree (event_type);


--
-- Name: idx_sendgrid_events_sg_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sendgrid_events_sg_event_id ON public.sendgrid_events USING btree (sg_event_id);


--
-- Name: idx_sendgrid_events_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sendgrid_events_user_id ON public.sendgrid_events USING btree (user_id);


--
-- Name: idx_socketlabs_events_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_socketlabs_events_user_id ON public.socketlabs_events USING btree (user_id);


--
-- Name: idx_sparkpost_events_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sparkpost_events_user_id ON public.sparkpost_events USING btree (user_id);


--
-- Name: postmark_events update_postmark_event_counts; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_postmark_event_counts AFTER INSERT ON public.postmark_events FOR EACH ROW EXECUTE FUNCTION public.update_event_counts();


--
-- Name: sendgrid_events update_sendgrid_event_counts; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_sendgrid_event_counts AFTER INSERT ON public.sendgrid_events FOR EACH ROW EXECUTE FUNCTION public.update_event_counts();


--
-- Name: socketlabs_events update_socketlabs_event_counts; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_socketlabs_event_counts AFTER INSERT ON public.socketlabs_events FOR EACH ROW EXECUTE FUNCTION public.update_event_counts();


--
-- Name: sparkpost_events update_sparkpost_event_counts; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_sparkpost_event_counts AFTER INSERT ON public.sparkpost_events FOR EACH ROW EXECUTE FUNCTION public.update_event_counts();


--
-- Name: email_service_providers email_service_providers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT email_service_providers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: sendgrid_events fk_sendgrid_events_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sendgrid_events
    ADD CONSTRAINT fk_sendgrid_events_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: socketlabs_events fk_socketlabs_events_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.socketlabs_events
    ADD CONSTRAINT fk_socketlabs_events_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: sparkpost_events fk_sparkpost_events_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sparkpost_events
    ADD CONSTRAINT fk_sparkpost_events_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: email_service_providers fk_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

