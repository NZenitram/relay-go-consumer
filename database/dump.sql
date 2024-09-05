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
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    id integer NOT NULL,
    message_id character varying(255) NOT NULL,
    processed boolean DEFAULT false,
    processed_time bigint,
    delivered boolean DEFAULT false,
    delivered_time bigint,
    bounce boolean DEFAULT false,
    bounce_type character varying(100),
    bounce_time bigint,
    deferred boolean DEFAULT false,
    deferred_count integer DEFAULT 0,
    last_deferral_time bigint,
    unique_open boolean DEFAULT false,
    unique_open_time bigint,
    open boolean DEFAULT false,
    open_count integer DEFAULT 0,
    last_open_time bigint,
    dropped boolean DEFAULT false,
    dropped_time bigint,
    dropped_reason text,
    provider character varying(25),
    metadata jsonb
);


--
-- Name: events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.events_id_seq OWNED BY public.events.id;


--
-- Name: message_user_associations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.message_user_associations (
    id integer NOT NULL,
    message_id character varying(255) NOT NULL,
    user_id integer NOT NULL,
    esp_id integer NOT NULL,
    provider character varying(25) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: message_user_associations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.message_user_associations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: message_user_associations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.message_user_associations_id_seq OWNED BY public.message_user_associations.id;


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
-- Name: events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events ALTER COLUMN id SET DEFAULT nextval('public.events_id_seq'::regclass);


--
-- Name: message_user_associations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_user_associations ALTER COLUMN id SET DEFAULT nextval('public.message_user_associations_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- Data for Name: email_service_providers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.email_service_providers (esp_id, user_id, provider_name, sending_domains, created_at, updated_at, sendgrid_verification_key, sparkpost_webhook_user, sparkpost_webhook_password, socketlabs_secret_key, postmark_webhook_user, postmark_webhook_password, socketlabs_server_id) FROM stdin;
5	5	sendgrid	{nzenitram.com,esprelay.com,webhookrelays.com,clickaleague.com}	2024-09-02 17:29:54.872196	2024-09-02 17:29:54.872196	vAzMaMUvUkR62eYW5cv7zgts8UHJTxe7r7rhTV4jqbnsPqqjABXzE63cUyxXVbYCU9okN1XXcr6G67Huii	\N	\N	\N	\N	\N	\N
4	4	postmark	{nzenitram.com,esprelay.com,webhookrelays.com,clickaleague.com}	2024-09-02 17:29:54.871366	2024-09-02 17:29:54.871366	\N	\N	\N	\N	user_id_four	FEfAtokCnyPeoL8	\N
2	2	sparkpost	{nzenitram.com,esprelay.com,webhookrelays.com,clickaleague.com}	2024-09-02 17:29:54.868415	2024-09-02 17:29:54.868415	\N	user_id_two	y7ELqPpLH5DHbt6	\N	\N	\N	\N
7	5	socketlabs	{nzenitram.com,esprelay.com,webhookrelays.com,clickaleague.com}	2024-09-02 17:29:54.873614	2024-09-02 17:29:54.873614	\N	\N	\N	j1qh04m0wH001jf	\N	\N	39044
3	3	socketlabs	{nzenitram.com,esprelay.com,webhookrelays.com,clickaleague.com}	2024-09-02 17:29:54.870227	2024-09-02 17:29:54.870227	\N	\N	\N	p7P8GoDk2r4KAt69	\N	\N	39044
6	5	sparkpost	{nzenitram.com,esprelay.com,webhookrelays.com,clickaleague.com}	2024-09-02 17:29:54.872812	2024-09-02 17:29:54.872812	\N	rg-postmark-user	KfR1U4YVpHAuq9o	\N	\N	\N	\N
1	1	sendgrid	{nzenitram.com,esprelay.com,webhookrelays.com,clickaleague.com}	2024-09-02 17:29:54.854658	2024-09-02 17:29:54.854658	MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEaku6twi1L2NsINLGt6fqGZWrVUU5mzhrep5ogR4z4WJcNmsZ4gKUULS/CatYYd+tgPrShNcYH9aKg1f0NTHgwA==	\N	\N	\N	\N	\N	\N
8	5	postmark	{nzenitram.com,esprelay.com,webhookrelays.com,clickaleague.com}	2024-09-02 17:29:54.874087	2024-09-02 17:29:54.874087	\N	\N	\N	\N	postmark_webhook_user	MT5X2e05uvmuQ28	\N
\.


--
-- Data for Name: events; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.events (id, message_id, processed, processed_time, delivered, delivered_time, bounce, bounce_type, bounce_time, deferred, deferred_count, last_deferral_time, unique_open, unique_open_time, open, open_count, last_open_time, dropped, dropped_time, dropped_reason, provider, metadata) FROM stdin;
2	Lot8rzLcTueVhDYjd6XiCA.recvd-6b556c7f5c-7qvdm-1-66D64F6F-7.2	t	1725321079	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	f	\N	\N	sendgrid	\N
3	Lot8rzLcTueVhDYjd6XiCA.recvd-6b556c7f5c-7qvdm-1-66D64F6F-7.0	t	1725321079	f	1725321073	f		\N	f	0	\N	f	\N	f	0	\N	f	\N	\N	sendgrid	\N
1	Lot8rzLcTueVhDYjd6XiCA.recvd-6b556c7f5c-7qvdm-1-66D64F6F-7.1	t	1725321150	f	1725321073	f		\N	f	0	\N	t	1725321133	t	1	1725321133	f	\N	\N	sendgrid	\N
4	tYh59iyqTveDX3H8UTZ4ng.recvd-57c9cdb5d8-2rqxw-1-66D65017-1.2	t	1725321239	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	f	\N	\N	sendgrid	\N
5	tYh59iyqTveDX3H8UTZ4ng.recvd-57c9cdb5d8-2rqxw-1-66D65017-1.0	t	1725321249	t	1725321241	f		\N	f	0	\N	f	\N	f	0	\N	f	\N	\N	sendgrid	\N
16	53c16815-6b27-438a-8a51-929d1c7d7941	t	1725375819	f	\N	t	HardBounce	1725332188	f	0	\N	f	\N	f	0	\N	t	1725332188	smtp;550 5.1.1 <tw1@nzenitram.com>: Recipient address rejected: User unknown in virtual mailbox table	postmark	\N
23	00000000-0000-0000-0000-000000000000	t	1725375921	f	\N	t	HardBounce	1725137930	f	0	\N	f	\N	f	0	\N	t	1725137930	Test bounce details	postmark	\N
9	qr1-guICR5SfqlVzVD85Cg.recvd-649fff897d-qdvjx-1-66D6646F-1A.2	t	1725326457	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	t	1725326448	Bounced Address	sendgrid	\N
8	qr1-guICR5SfqlVzVD85Cg.recvd-649fff897d-qdvjx-1-66D6646F-1A.1	t	1725326457	f	1725326449	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		sendgrid	\N
7	qr1-guICR5SfqlVzVD85Cg.recvd-649fff897d-qdvjx-1-66D6646F-1A.0	t	1725326500	f	1725326449	f		\N	f	0	\N	t	1725326489	t	1	1725326489	f	\N		sendgrid	\N
31	8wCoHm0RTkaAwWlyfPSdYw.recvd-6b556c7f5c-47rpr-1-66D7B383-2.1	t	1725412531	f	1725412229	f		\N	f	0	\N	f	1725412467	f	1	1725412467	f	\N		sendgrid	\N
24	c6168346-ea2c-4919-9bf1-788300623150	t	1725375942	f	\N	f		\N	f	0	\N	f	1725332104	t	1	\N	f	\N		postmark	\N
10	3c742836-edc8-491a-a730-f4469509a4ad	t	0	f	1725330586	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		postmark	\N
11	30fbe4d8-3b55-4405-a211-e88b7be1097c	t	0	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		postmark	\N
33	8wCoHm0RTkaAwWlyfPSdYw.recvd-6b556c7f5c-47rpr-1-66D7B383-2.2	t	1725412535	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	t	1725412227	Bounced Address	sendgrid	\N
12	502e4632-8195-4c62-98f5-314735777d2f	t	0	f	1725330794	f		\N	f	0	\N	t	\N	f	0	\N	f	\N		postmark	\N
21	66cfa78ad66631516650	t	1725336233	f	\N	t	550 5.1.1 ...@... Recipient address rejected: ... in virtual mailbox table	1725336233	f	0	\N	f	\N	f	0	\N	f	\N		sparkpost	\N
13	feb78ad2-9ef9-4c04-9ddd-473397c428cb	t	0	t	1725331191	f		1725331190	f	0	\N	f	\N	f	0	\N	f	1725331190		postmark	\N
14	71315168-21b0-4c07-b3fd-ce3ffcc24135	t	0	t	1725332038	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		postmark	\N
25	66cfa78ad66631516651	t	1725336233	f	\N	t	550 5.1.1 ...@... Recipient address rejected: ... in virtual mailbox table	1725336233	f	0	\N	f	\N	f	0	\N	f	\N		sparkpost	\N
17	cDdQOEdvRGsycjRLQXQ2OTozOTA0NDoxNzI1MzM0MDkyNDU5ODc1MDAw	t	1725251692	f	\N	f		\N	f	0	\N	t	1725251692	t	1	1725251692	f	\N		socketlabs	\N
27	f03GI5ndROGAOBvtLhjk5g.recvd-764b8dc56-hwwph-1-66D79B8D-9.0	t	1725412808	f	\N	f		\N	f	0	\N	t	1725411616	t	4	1725411616	f	\N		sendgrid	\N
19	66cf013ad566575300f7	t	1725250051	t	1725250051	f		\N	f	0	\N	f	1725250055	f	1	1725250055	f	\N		sparkpost	\N
18	SampleMessageId	t	1725145478	f	\N	f		\N	t	2	1725145478	f	\N	f	0	\N	f	\N	{"code":9999,"reason":"Sample Reason"}	socketlabs	\N
26	1234-5678-9012	t	1725145478	f	\N	f		\N	t	1	1725145478	f	\N	f	0	\N	f	\N	{"code":9999,"reason":"Sample Reason"}	socketlabs	\N
6	tYh59iyqTveDX3H8UTZ4ng.recvd-57c9cdb5d8-2rqxw-1-66D65017-1.1	t	1725326524	f	1725321241	f		\N	f	0	\N	t	1725326517	t	3	1725326517	f	\N		sendgrid	\N
22	66cf6e8ed66623265a57	t	1725337251	f	1725337202	f		\N	f	0	\N	t	1725337251	t	1	1725337251	f	\N		sparkpost	\N
36	2V_BfG9qTyqtl1mIpDE2sg.recvd-764b8dc56-hwwph-1-66D7B5F9-15.0	t	1725413182	f	1725412878	f		\N	f	0	\N	t	1725413179	t	1	1725413179	f	\N		sendgrid	\N
30	D5sarl2zSseyXoU74KMIMg.recvd-6b556c7f5c-jjv9p-1-66D7B30C-3.2	t	1725412109	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	t	1725412108	Bounced Address	sendgrid	\N
28	D5sarl2zSseyXoU74KMIMg.recvd-6b556c7f5c-jjv9p-1-66D7B30C-3.0	t	1725412119	t	1725412110	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		sendgrid	\N
29	D5sarl2zSseyXoU74KMIMg.recvd-6b556c7f5c-jjv9p-1-66D7B30C-3.1	t	1725412481	f	1725412110	f		\N	f	0	\N	t	1725412465	t	1	1725412465	f	\N		sendgrid	\N
32	8wCoHm0RTkaAwWlyfPSdYw.recvd-6b556c7f5c-47rpr-1-66D7B383-2.0	t	1725412531	f	1725412229	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		sendgrid	\N
34	2V_BfG9qTyqtl1mIpDE2sg.recvd-764b8dc56-hwwph-1-66D7B5F9-15.1	t	1725412880	f	1725412878	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		sendgrid	\N
35	2V_BfG9qTyqtl1mIpDE2sg.recvd-764b8dc56-hwwph-1-66D7B5F9-15.2	t	1725412875	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	t	1725412877	Bounced Address	sendgrid	\N
39	vJeX94XGStS2CMp49dnPgw.recvd-57c9cdb5d8-sk577-1-66D7B679-B.2	t	1725412989	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	t	1725412985	Bounced Address	sendgrid	\N
38	vJeX94XGStS2CMp49dnPgw.recvd-57c9cdb5d8-sk577-1-66D7B679-B.1	t	1725412989	f	1725412987	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		sendgrid	\N
37	vJeX94XGStS2CMp49dnPgw.recvd-57c9cdb5d8-sk577-1-66D7B679-B.0	t	1725413191	f	1725412987	f		\N	f	0	\N	t	1725413178	t	1	1725413178	f	\N		sendgrid	\N
40	FLjX_-2nSLWmBfWneuWcDg.recvd-84679d98c6-8k98j-1-66D7B7BF-B.0	t	1725413360	f	1725413313	f		\N	f	0	\N	t	1725413357	t	1	1725413357	f	\N		sendgrid	\N
41	cDdQOEdvRGsycjRLQXQ2OTozOTA0NDoxNzI1NDE3NjY2NDE2MDM2MDAw	t	1725417660	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		socketlabs	\N
42	cDdQOEdvRGsycjRLQXQ2OTozOTA0NDoxNzI1NDE3NjcwMzY4MzY1MDAw	t	1725417667	t	1725417667	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		socketlabs	\N
43	cDdQOEdvRGsycjRLQXQ2OTozOTA0NDoxNzI1NDE3ODQyNDUwODIxMDAw	t	1725417724	f	\N	f		\N	f	0	\N	t	1725417724	t	1	1725417724	f	\N		socketlabs	\N
44	f9186fdc-8bec-4c5c-8751-a41a1de6bd68	t	1725418402	f	1725417980	f		\N	f	0	\N	t	1725418399	t	1	\N	f	\N		postmark	\N
20	000443ee14578172be22	t	1460989507	f	1460989507	f		1460989507	f	1	1460989507	f	1460989507	f	1	1460989507	f	1460989507		sparkpost	\N
45		f	0	f	\N	f		\N	f	0	\N	f	\N	f	0	\N	f	\N			\N
46	66d7ebd2d76625353f1f	t	1725420269	t	1725420269	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		sparkpost	\N
47	66cfa4d4d7665933598d	t	1725420711	t	1725420711	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		sparkpost	\N
48	66cff1d1d76654366ad6	t	1725420019	t	1725420019	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		sparkpost	\N
49	66cfa1cfd7664d3c6c8d	t	1725419427	t	1725419427	f		\N	f	0	\N	f	\N	f	0	\N	f	\N		sparkpost	\N
\.


--
-- Data for Name: message_user_associations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.message_user_associations (id, message_id, user_id, esp_id, provider, created_at) FROM stdin;
11	FLjX_-2nSLWmBfWneuWcDg.recvd-84679d98c6-8k98j-1-66D7B7BF-B.0	1	1	sendgrid	2024-09-04 01:28:40.336507
14	f9186fdc-8bec-4c5c-8751-a41a1de6bd68	4	4	postmark	2024-09-04 02:53:22.419969
15		2	2	sparkpost	2024-09-04 03:24:51.164572
17	66cfa4d4d7665933598d	2	2	sparkpost	2024-09-04 03:38:10.370048
21	66cff1d1d76654366ad6	2	2	sparkpost	2024-09-04 03:45:12.575729
22	66cfa1cfd7664d3c6c8d	2	2	sparkpost	2024-09-04 03:47:16.473752
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, username, email, first_name, last_name, created_at, updated_at, api_key) FROM stdin;
1	johndoe	john.doe@example.com	John	Doe	2024-09-02 16:11:35.434317	2024-09-02 16:11:35.434317	q6u0Noz6CbTvj7LyTYIblHL9sklPs7wV
2	janesmit	jane.smith@example.com	Jane	Smith	2024-09-02 16:11:35.434317	2024-09-02 16:11:35.434317	JwJ1XmqR7lG2Xoo9yqXD2OchzghFXGtc
3	bobwilson	bob.wilson@example.com	Bob	Wilson	2024-09-02 16:11:35.434317	2024-09-02 16:11:35.434317	Ttv5ISk5gMJiED6kq3sDXJZYUvOwv5we
4	alicecooper	alice.cooper@example.com	Alice	Cooper	2024-09-02 17:24:25.28577	2024-09-02 17:24:25.28577	a1b2c3d4e5f6g7h8i9
5	bobmarley	bob.marley@example.com	Bob	Marley	2024-09-02 17:24:25.28577	2024-09-02 17:24:25.28577	j1k2l3m4n5o6p7q8r9
\.


--
-- Name: email_service_providers_esp_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.email_service_providers_esp_id_seq', 1, false);


--
-- Name: events_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.events_id_seq', 49, true);


--
-- Name: message_user_associations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.message_user_associations_id_seq', 24, true);


--
-- Name: users_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.users_user_id_seq', 5, true);


--
-- Name: email_service_providers email_service_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT email_service_providers_pkey PRIMARY KEY (esp_id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: message_user_associations message_user_associations_message_id_provider_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_user_associations
    ADD CONSTRAINT message_user_associations_message_id_provider_key UNIQUE (message_id, provider);


--
-- Name: message_user_associations message_user_associations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_user_associations
    ADD CONSTRAINT message_user_associations_pkey PRIMARY KEY (id);


--
-- Name: email_service_providers unique_postmark_password; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT unique_postmark_password UNIQUE (postmark_webhook_password);


--
-- Name: email_service_providers unique_sendgrid_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT unique_sendgrid_key UNIQUE (sendgrid_verification_key);


--
-- Name: email_service_providers unique_socketlabs_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT unique_socketlabs_key UNIQUE (socketlabs_secret_key);


--
-- Name: email_service_providers unique_sparkpost_password; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT unique_sparkpost_password UNIQUE (sparkpost_webhook_password);


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
-- Name: idx_events_bounced; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_bounced ON public.events USING btree (bounce_time) WHERE (bounce = true);


--
-- Name: idx_events_delivered; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_delivered ON public.events USING btree (delivered_time) WHERE (delivered = true);


--
-- Name: idx_events_message_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_events_message_id ON public.events USING btree (message_id);


--
-- Name: idx_events_message_id_provider; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_message_id_provider ON public.events USING btree (message_id, provider);


--
-- Name: idx_events_opened; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_opened ON public.events USING btree (last_open_time) WHERE (open = true);


--
-- Name: idx_events_processed_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_processed_time ON public.events USING btree (processed_time);


--
-- Name: idx_events_provider; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_provider ON public.events USING btree (provider);


--
-- Name: idx_message_user_associations_esp_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_message_user_associations_esp_id ON public.message_user_associations USING btree (esp_id);


--
-- Name: idx_message_user_associations_message_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_message_user_associations_message_id ON public.message_user_associations USING btree (message_id);


--
-- Name: idx_message_user_associations_provider; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_message_user_associations_provider ON public.message_user_associations USING btree (provider);


--
-- Name: idx_message_user_associations_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_message_user_associations_user_id ON public.message_user_associations USING btree (user_id);


--
-- Name: email_service_providers email_service_providers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT email_service_providers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: email_service_providers fk_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_service_providers
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: message_user_associations message_user_associations_esp_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_user_associations
    ADD CONSTRAINT message_user_associations_esp_id_fkey FOREIGN KEY (esp_id) REFERENCES public.email_service_providers(esp_id);


--
-- Name: message_user_associations message_user_associations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_user_associations
    ADD CONSTRAINT message_user_associations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

