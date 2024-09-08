Certainly! I'll update the README to include information about using an authorization header with the user's API key. Here's the revised structure with the new section added:

# Email Sending Service Documentation

## Table of Contents
1. [Introduction](#introduction)
2. [API Endpoint](#api-endpoint)
3. [Authentication](#authentication)
4. [JSON Payload Structure](#json-payload-structure)
5. [Personalization and Content](#personalization-and-content)
6. [Attachments](#attachments)
7. [Custom Headers and Tracking](#custom-headers-and-tracking)
8. [Email Service Provider Credentials](#email-service-provider-credentials)
9. [Example CURL Requests](#example-curl-requests)
10. [Webhook Events](#webhook-events)
11. [Troubleshooting](#troubleshooting)

## Introduction
This document outlines the structure and usage of our email sending service, which supports multiple email service providers and offers personalization features.

## API Endpoint
The email sending endpoint is accessible at:
```
https://horribly-striking-joey.ngrok-free.app/emails
```

## Authentication
To authorize email sends, you must include your API key in the Authorization header of your HTTP request.

Header format:
```
Authorization: Bearer YOUR_API_KEY
```

Replace `YOUR_API_KEY` with the actual API key provided to you.

## JSON Payload Structure
The API accepts a JSON payload with the following main sections:
- `from`: Sender information
- `personalizations`: Recipient-specific information and substitutions
- `content`: Email body in plain text and HTML formats
- `sections`: Reusable content sections
- `attachments`: File attachments
- `headers`: Custom email headers
- `custom_args`: Additional sending options
- `categories`: Email categorization

## Personalization and Content
### Personalizations
The `personalizations` array allows for recipient-specific customization:
```json
"personalizations": [
  {
    "to": {
      "name": "Recipient Name",
      "email": "recipient@example.com"
    },
    "subject": "Personalized Subject",
    "substitutions": {
      "name": "Recipient Name",
      "order_id": "12345",
      "confirmations": "confirmation_001"
    }
  }
]
```

### Content
Email content can be provided in both plain text and HTML formats. Variables can be included using either `-var-` or `{{var}}` syntax:
```json
"content": [
  {
    "type": "text/plain",
    "value": "Hello {{name}},\n{{confirmations}}"
  },
  {
    "type": "text/html",
    "value": "<html><body><p>Hello {{name}},<br>{{confirmations}}</p></body></html>"
  }
]
```

### Sections
Reusable content sections can be defined and referenced in the email content:
```json
"sections": {
  "confirmation_001": "Thanks for choosing our service. This email is to confirm that we have processed your order {{order_id}}.",
  "confirmation_002": "Thanks for your order. We've processed your order {{order_id}}. You can download your invoice as a PDF for your records."
}
```

## Attachments
Files can be attached to the email:
```json
"attachments": [
  {
    "filename": "example.txt",
    "type": "text/plain",
    "content": "SGVsbG8gd29ybGQh",
    "content_id": "ii_139db99fdb5c3704",
    "disposition": "attachment"
  }
]
```

## Custom Headers and Tracking
Custom headers and tracking options can be specified:
```json
"headers": {
  "X-Custom-Header-1": "Custom Value 1",
  "X-Custom-Header-2": "Custom Value 2"
},
"custom_args": {
  "TrackOpens": "true",
  "TrackLinks": "HtmlOnly",
  "MessageStream": "outbound"
}
```

## Email Service Provider Credentials
The API supports multiple email service providers. Credentials and weights for each provider can be specified in the request payload or managed server-side.

## Example CURL Requests
Here's an example CURL request to send an email:
```bash
curl -X POST https://horribly-striking-joey.ngrok-free.app/emails \
-H "Content-Type: application/json" \
-H "Authorization: Bearer YOUR_API_KEY" \
-d '{
  "from": {
    "name": "Sender Name",
    "email": "sender@example.com"
  },
  "personalizations": [
    {
      "to": {
        "name": "Recipient Name",
        "email": "recipient@example.com"
      },
      "subject": "Personalized Subject",
      "substitutions": {
        "name": "Recipient Name",
        "order_id": "12345",
        "confirmations": "confirmation_001"
      }
    }
  ],
  "content": [
    {
      "type": "text/plain",
      "value": "Hello {{name}},\n{{confirmations}}"
    },
    {
      "type": "text/html",
      "value": "<html><body><p>Hello {{name}},<br>{{confirmations}}</p></body></html>"
    }
  ],
  "sections": {
    "confirmation_001": "Thanks for choosing our service. This email is to confirm that we have processed your order {{order_id}}."
  },
  "attachments": [
    {
      "filename": "receipt.pdf",
      "type": "application/pdf",
      "content": "base64encodedcontent"
    }
  ],
  "headers": {
    "X-Custom-Header": "Custom Value"
  },
  "custom_args": {
    "TrackOpens": "true",
    "TrackLinks": "HtmlOnly"
  },
  "categories": ["order", "confirmation"]
}'
```

## Webhook Events
The service supports webhook events for tracking email status. Webhook endpoints for different providers are available for processing these events.

## Troubleshooting
For any issues, please refer to the specific email service provider's documentation or contact our support team.
