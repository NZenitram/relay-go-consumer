# Email Processing System (Consumer)

This is the consumer component of our Go-based email processing system. It works in tandem with the ingress system to process and send emails, as well as handle webhook events from various Email Service Providers (ESPs).

## Overview

The consumer system is responsible for:
1. Consuming messages from Kafka topics
2. Processing email send requests
3. Handling webhook events from different ESPs (SendGrid, Postmark, SocketLabs, SparkPost)
4. Managing email delivery through multiple ESPs using a weighted approach
5. Error handling and retries
6. Updating the database with email statuses and event data

## Key Components

1. **Main Application (main.go)**: Sets up Kafka consumers for various topics and manages the overall flow of the application.

2. **Email Processor**: Handles the core email processing logic, including batch processing and ESP weighting.

3. **ESP Integrations**: Separate modules for each supported ESP (SendGrid, Postmark, SocketLabs, SparkPost) to handle sending emails and processing webhook events.

4. **Webhook Event Processor**: Manages incoming webhook events from different ESPs.

5. **Database Integration**: Handles database operations for storing and retrieving email and event data.

## Email Processing and Sending

The system uses a sophisticated approach to process and send emails:

1. **Batch Processing**: Emails can be processed in batches, allowing for efficient handling of large volumes.

2. **ESP Weighting**: The system calculates weights for each ESP based on their performance over time. This includes factors such as:
   - Open rates
   - Delivery rates
   - Bounce rates
   - Spam report rates

3. **Dynamic Weight Calculation**: 
   - For the first batch or non-batch emails, it uses data from the last 30 days.
   - For subsequent batches, it uses data since the last batch was sent.
   - Weights are normalized to sum up to 1000 for precise distribution.

4. **Sender Selection**: Based on the calculated weights, the system selects an appropriate ESP for each email or group of emails.

5. **Personalization**: The system supports personalized emails, using substitutions provided in the email payload.

6. **Multi-ESP Sending**: Emails within a single request can be distributed across multiple ESPs based on their weights.

## Event Processing

The system processes various types of events from different ESPs:

1. **Webhook Consumption**: Dedicated Kafka topics for each ESP's webhook events.

2. **Event Types**: Processes events such as deliveries, opens, clicks, bounces, and spam reports.

3. **Database Updates**: Each processed event updates the relevant email status in the database.

4. **Performance Tracking**: Event data is used to calculate ESP performance metrics, which in turn affects future weight calculations.

## Configuration

The application uses environment variables for configuration. Key variables include:

- `KAFKA_BROKERS`: Kafka broker addresses
- `KAFKA_EMAIL_TOPIC`: Topic for email messages
- `WEBHOOK_TOPIC_*`: Topics for webhook events from different ESPs
- `KAFKA_OFFSET_RESET`: Kafka consumer offset reset policy

## Running the Application

1. Ensure all environment variables are set in the `.env` file.
2. Build the Docker image: `docker build -t email-consumer .`
3. Run the container: `docker run --env-file .env email-consumer`

## Database Seeding

The application includes a database seeding option for development and testing purposes. To seed the database:

```
go run main.go -seed
```

## Error Handling

Each ESP integration includes specific error handling logic to manage API errors and retry mechanisms.

## Performance Considerations

- The application uses goroutines to consume messages from different Kafka topics concurrently.
- ESP weighting helps in load balancing and optimizing email delivery across multiple providers.
- Batch processing is implemented for efficient handling of large volumes of emails.

## Contributing

Please refer to CONTRIBUTING.md for guidelines on how to contribute to this project.

## License

[Specify your license here]
