## Email Sending Endpoint JSON Structure

This section describes the JSON structure used to send an email through the email sending endpoint. The JSON object contains various fields that specify the sender, recipients, email content, additional options, and credentials for API authentication.

### JSON Fields

- **`from`**: A string representing the sender's email address and name. It follows the format `"Name <email@example.com>"`.

- **`to`**: An array of strings, each representing a recipient's email address and name. Each entry follows the format `"Name <email@example.com>"`.

- **`cc`**: An array of strings, each representing a CC (carbon copy) recipient's email address. This field is optional.

- **`bcc`**: An array of strings, each representing a BCC (blind carbon copy) recipient's email address. This field is optional.

- **`subject`**: A string representing the subject line of the email.

- **`textbody`**: A string containing the plain text version of the email content. This version is used by email clients that do not support HTML.

- **`htmlbody`**: A string containing the HTML version of the email content. This version allows for rich text formatting and styling.

- **`attachments`**: An array of objects, each representing an attachment. Each attachment object contains:
  - **`name`**: A string specifying the file name of the attachment.
  - **`contenttype`**: A string specifying the MIME type of the attachment.
  - **`content`**: A base64-encoded string representing the content of the attachment.

- **`headers`**: An object containing custom email headers. Each key-value pair represents a header name and its corresponding value.

- **`data`**: An object containing additional email sending options:
  - **`TrackOpens`**: A boolean indicating whether to track email opens.
  - **`TrackLinks`**: A string specifying link tracking options. Possible values include `"HtmlOnly"`.
  - **`MessageStream`**: A string specifying the message stream to use, such as `"outbound"`.

- **`credentials`**: An object containing authentication credentials and configuration for sending emails through different services:
  - **`SocketLabsServerID`**: A string representing the server ID for SocketLabs.
  - **`SocketLabsAPIkey`**: A string representing the API key for SocketLabs.
  - **`SocketLabsWeight`**: A string representing the weighting factor for selecting SocketLabs as the email service.
  - **`PostmarkServerToken`**: A string representing the server token for Postmark.
  - **`PostmarkAPIURL`**: A string representing the API URL for Postmark.
  - **`PostmarkWeight`**: A string representing the weighting factor for selecting Postmark as the email service.

This JSON structure allows for detailed customization of the email sending process, including specifying recipients, adding attachments, setting tracking options, and using specific credentials for different email service providers.

## Example CURL Requests for Testing

### Sample JSON Structure for Email Sends

```json
{
  "from": {
    "name": "Twitter Zen",
    "email": "test@nzenitram.com"
  },
  "personalizations": [
    {
      "to": {
        "name": "Nick Martinez, Jr.",
        "email": "twitter1@nzenitram.com"
      },
      "subject": "Personalized Email for Nick",
      "substitutions": {
        "-name-": "Nick",
        "-order_id-": "12345",
        "-confirmations-": "-confirmation_001-"
      }
    },
    {
      "to": {
        "name": "Jane Doe",
        "email": "jane@example.com"
      },
      "subject": "Personalized Email for Jane",
      "substitutions": {
        "-name-": "Jane",
        "-order_id-": "67890",
        "-confirmations-": "-confirmation_002-"
      }
    }
  ],
  "cc": ["nick1@nzenitram.com"],
  "bcc": ["support@nzenitram.com"],
  "subject": "Default Subject (if no personalization)",
  "content": [
    {
      "type": "text/plain",
      "value": "Hello -name-,\n-confirmations-"
    },
    {
      "type": "text/html",
      "value": "<html><head></head><body><p>Hello -name-,<br>-confirmations-</p></body></html>"
    }
  ],
  "sections": {
    "-confirmation_001-": "Thanks for choosing our service. This email is to confirm that we have processed your order -order_id-.",
    "-confirmation_002-": "Thanks for your order. We'\''ve processed your order -order_id-. You can download your invoice as a PDF for your records."
  },
  "attachments": [
    {
      "filename": "example.txt",
      "type": "text/plain",
      "content": "SGVsbG8gd29ybGQh"
    }
  ],
  "headers": {
    "X-Custom-Header-1": "Custom Value 1",
    "X-Custom-Header-2": "Custom Value 2"
  },
  "custom_args": {
    "TrackOpens": "true",
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
  },
  "categories": ["test", "example"]
}
```

```bash
curl -X POST https://horribly-striking-joey.ngrok-free.app/emails -H "Content-Type: application/json" -d '{
  "from": {
    "name": "Twitter Zen",
    "email": "admin@esprelay.com"
  },
  "personalizations": [
    {
      "to": {
        "name": "Nick Martinez, Jr.",
        "email": "twitter1@nzenitram.com"
      },
      "subject": "Personalized Email for Nick",
      "substitutions": {
        "-name-": "Nick",
        "-order_id-": "12345",
        "-confirmations-": "This is Confirmaton 001"
      }
    },
    {
      "to": {
        "name": "Jane Doe",
        "email": "nzenitram@nzenitram.com"
      },
      "subject": "Personalized Email for Jane",
      "substitutions": {
        "-name-": "Jane",
        "-order_id-": "67890",
        "-confirmations-": "This is Confirmaton 002"
      }
    }
  ],
  "cc": ["nick1@nzenitram.com"],
  "bcc": ["support@nzenitram.com"],
  "subject": "Default Subject (if no personalization)",
  "content": [
    {
      "type": "text/plain",
      "value": "Hello -name-,\n-confirmations-"
    },
    {
      "type": "text/html",
      "value": "<html><head></head><body><p>Hello -name-,<br>-confirmations-</p></body></html>"
    }
  ],
  "sections": {
    "-confirmation_001-": "Thanks for choosing our service. This email is to confirm that we have processed your order -order_id-.",
    "-confirmation_002-": "Thanks for your order. We'\''ve processed your order -order_id-. You can download your invoice as a PDF for your records."
  },
  "attachments": [
    {
      "filename": "example.txt",
      "type": "text/plain",
      "content": "SGVsbG8gd29ybGQh",
      "content_id": "ii_139db99fdb5c3704",
      "name": "example",
      "disposition": "attachment"
    }
  ],
  "headers": {
    "X-Custom-Header-1": "Custom Value 1",
    "X-Custom-Header-2": "Custom Value 2"
  },
  "custom_args": {
    "TrackOpens": "true",
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
  },
  "categories": [
    "test", "example"
  ],
  "credentials": {
    "SocketLabsServerID": "39044",
    "SocketLabsAPIkey": "Jn42Qbx7H5TyFo36Wzp8",
    "SocketLabsWeight": "0",
    "PostmarkServerToken": "66422125-a3f0-4690-9279-2bdaf804f19f",
    "PostmarkWeight": "0",
    "SendgridAPIKey": "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",
    "SendgridWeight": "100",
    "SparkpostAPIKey": "d62547b5a928d878d08fdb6e17d682bf1d61c364",
    "SparkpostWeight": "0"
  }
}'
```

```json
{
  "from": "Twitter Zen <nick@nzenitram.com>",
  "to": [
    "\"Nick Martinez, Jr.\" <twitter1@nzenitram.com>",
    "Mick Nartinez <nzenitram@nzenitram.com>"
  ],
  "cc": ["nick1@nzenitram.com"],
  "bcc": ["support@nzenitram.com"],
  "subject": "Updating the subject to reflect the test",
  "textbody": "This is the plain text body of the email.",
  "htmlbody": "<p>This is the <strong>HTML</strong> body of the email.</p>",
  "attachments": [
    {
      "name": "example.txt",
      "contenttype": "text/plain",
      "content": "SGVsbG8gd29ybGQh"
    }
  ],
  "headers": {
    "X-Custom-Header-1": "Custom Value 1",
    "X-Custom-Header-2": "Custom Value 2"
  },
  "data": {
    "TrackOpens": true,
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
  },
  "credentials": {
    "SocketLabsServerID": "12345",
    "SocketLabsAPIkey": "12345abcdefg",
    "SocketLabsWeight": "50",
    "PostmarkServerToken": "555555555-abcd-5555-9279-2bdaf804f19f",
    "PostmarkWeight": "50"
  }
}
```

```json
{
	"headers": {
		"Accept": [
			"*/*"
		],
		"Accept-Encoding": [
			"gzip"
		],
		"Content-Length": [
			"1417"
		],
		"Content-Type": [
			"application/json"
		],
		"User-Agent": [
			"curl/8.7.1"
		],
		"X-Forwarded-For": [
			"67.1.195.103"
		],
		"X-Forwarded-Host": [
			"horribly-striking-joey.ngrok-free.app"
		],
		"X-Forwarded-Proto": [
			"https"
		]
	},
	"body": {
		"from": "Twitter Zen <test@esprelay.com>",
		"to": [
			"\"Nick Martinez, Jr.\" <twitter1@nzenitram.com>",
			"Mick Nartinez <nzenitram@nzenitram.com>",
			"Admin ESP <admin@esprelay.com>",
			"Admin Webhook <admin@webhookrelays.com>",
			"Test ESP <test@esprelay.com>",
			"YF CLickALeague <yourfriends@clickaleague.com>"
		],
		"cc": [
			"nick1@nzenitram.com"
		],
		"bcc": [
			"support@nzenitram.com"
		],
		"subject": "Updating the subject to reflect the test",
		"textbody": "",
		"htmlbody": "<p>This is the <strong>HTML</strong> body of the email.</p>",
		"attachments": [
			{
				"name": "example.txt",
				"contenttype": "text/plain",
				"content": "SGVsbG8gd29ybGQh"
			}
		],
		"headers": {
			"X-Custom-Header-1": "Custom Value 1",
			"X-Custom-Header-2": "Custom Value 2"
		},
		"data": {
			"TrackOpens": true,
			"TrackLinks": "HtmlOnly",
			"MessageStream": "outbound"
		},
		"credentials": {
    "SocketLabsServerID": "12345",
    "SocketLabsAPIkey": "12345abcdefg",
    "SocketLabsWeight": "25",
    "PostmarkServerToken": "555555555-abcd-5555-9279-2bdaf804f19f",
    "PostmarkWeight": "25",
    "SendgridAPIKey": "SG.asdfasfasdfasfd---asdf.x_x7thLJSL0muwesXAPVgZcgA0",
    "SendgridWeight": "25",
    "SparkpostAPIKey": "asdfasdf3224512345sadfasdf",
    "SparkpostWeight": "25"
		}
	}
}
```
### Curl with Text Body, Friendly From, Headers, and Attachment

```bash
curl -X POST http://localhost:8888 -H "Content-Type: application/json" -d '{
  "from": "Twitter Zen <nick@nzenitram.com>",
  "to": [
    "\"Nick Martinez, Jr.\" <nick@nzenitram.com>",
    "Mick Nartinez <nzenitram@nzenitram.com>"
  ],
  "cc": ["nick1@nzenitram.com"],
  "bcc": ["support@nzenitram.com"],
  "subject": "Updating the subject to reflect the test",
  "textbody": "This is the plain text body of the email.",
  "htmlbody": "",
  "attachments": [
    {
      "name": "example.txt",
      "contenttype": "text/plain",
      "content": "SGVsbG8gd29ybGQh"
    }
  ],
  "headers": {
    "X-Custom-Header-1": "Custom Value 1",
    "X-Custom-Header-2": "Custom Value 2"
  },
  "data": {
    "TrackOpens": true,
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
  },
  "credentials": {
    "SocketLabsServerID": "12345",
    "SocketLabsAPIkey": "12345abcdefg",
    "SocketLabsWeight": "50",
    "PostmarkServerToken": "555555555-abcd-5555-9279-2bdaf804f19f",
    "PostmarkWeight": "50"
  }
}'
```

### Curl with HTML Body, Friendly From, Headers, and Attachment

```bash
curl -X POST http://localhost:8888 -H "Content-Type: application/json" -d '{
  "from": "Twitter Zen <nick@nzenitram.com>",
  "to": [
    "\"Nick Martinez, Jr.\" <nick@nzenitram.com>",
    "Mick Nartinez <nzenitram@nzenitram.com>"
  ],
  "cc": ["nick1@nzenitram.com"],
  "bcc": ["support@nzenitram.com"],
  "subject": "Updating the subject to reflect the test",
  "textbody": "",
  "htmlbody": "<p>This is the <strong>HTML</strong> body of the email.</p>",
  "attachments": [
    {
      "name": "example.txt",
      "contenttype": "text/plain",
      "content": "SGVsbG8gd29ybGQh"
    }
  ],
  "headers": {
    "X-Custom-Header-1": "Custom Value 1",
    "X-Custom-Header-2": "Custom Value 2"
  },
  "data": {
    "TrackOpens": true,
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
  },
  "credentials": {
    "SocketLabsServerID": "12345",
    "SocketLabsAPIkey": "12345abcdefg",
    "SocketLabsWeight": "50",
    "PostmarkServerToken": "555555555-abcd-5555-9279-2bdaf804f19f",
s    "PostmarkWeight": "50"
  }
}'
```

```bash
curl -X POST https://horribly-striking-joey.ngrok-free.app/emails -H "Content-Type: application/json" -d  '{
  "from": "Twitter Zen <test@nzenitram.com>",
  "to": [
    "\"Nick Martinez, Jr.\" <twitter1@nzenitram.com>",
    "Mick Nartinez <nzenitram@nzenitram.com>",
    "Admin ESP <admin@esprelay.com>",
    "Admin Webhook <admin@webhookrelays.com>",
    "Test ESP <test@esprelay.com>",
    "YF CLickALeague <yourfriends@clickaleague.com>",
    "\"Nick Martinez, Jr.\" <twitter1@nzenitram.com>",
    "Mick Nartinez <nzenitram@nzenitram.com>",
    "Admin ESP <admin@esprelay.com>",
    "Admin Webhook <admin@webhookrelays.com>",
    "Test ESP <test@esprelay.com>"
  ],
  "cc": ["nick1@nzenitram.com"],
  "bcc": ["support@nzenitram.com"],
  "subject": "Updating the subject to reflect the test",
  "textbody": "",
  "htmlbody": "<p>This is the <strong>HTML</strong> body of the email.</p>",
  "attachments": [
    {
      "name": "example.txt",
      "contenttype": "text/plain",
      "content": "SGVsbG8gd29ybGQh"
    }
  ],
  "headers": {
    "X-Custom-Header-1": "Custom Value 1",
    "X-Custom-Header-2": "Custom Value 2"
  },
  "data": {
    "TrackOpens": true,
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
    },
  "credentials": {
    "SocketLabsServerID": "39044",
    "SocketLabsAPIkey": "Jn42Qbx7H5TyFo36Wzp8",
    "SocketLabsWeight": "25",
    "PostmarkServerToken": "66422125-a3f0-4690-9279-2bdaf804f19f",
    "PostmarkWeight": "25",
    "SendgridAPIKey": "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",
    "SendgridWeight": "25",
    "SparkpostAPIKey": "d62547b5a928d878d08fdb6e17d682bf1d61c364",
    "SparkpostWeight": "25"
    }
}'
```



### Instructions for `textbody` vs `htmlbody`

- **`textbody`**: Use this field to provide a plain text version of the email content. This is important for recipients whose email clients do not support HTML. It ensures that the message is still readable without formatting.

- **`htmlbody`**: Use this field to provide an HTML version of the email content. This allows for rich text formatting, including styles, links, and images, enhancing the visual appeal of the email.

When both `textbody` and `htmlbody` are provided, email clients that support HTML will display the HTML content, while those that do not will fall back to the plain text content. This ensures broad compatibility and a good user experience across different email platforms.

### Using Credentials in API Calls

The `credentials` field in the JSON structure provides the necessary authentication details for sending emails through different services. Depending on the service being used (SocketLabs or Postmark), the relevant credentials and weights are utilized to authenticate and prioritize the email sending process. The weights determine the probability of selecting a particular service when sending emails, allowing for load balancing or preference-based sending strategies.

## Postmark API Error Code 412

When interacting with the Postmark API, you may encounter **Error Code 412**. This error occurs under specific conditions related to account approval status and domain restrictions. Below is an explanation of this error code and how to resolve it:

### **Error Code 412: Pending Approval Domain Restriction**

- **Description**: This error indicates that your Postmark account is still pending approval. During this period, there is a restriction that requires all recipient email addresses to share the same domain as the 'From' address used in your emails.
  
- **Example Scenario**: If your 'From' address is `twitter1@nzenitram.com`, you are only allowed to send emails to recipients with email addresses that also belong to the `nzenitram.com` domain. Attempting to send emails to other domains, such as `clickaleague.com`, will trigger this error.

- **Resolution**:
  - Ensure that all recipient addresses have the same domain as the 'From' address while your account is pending approval.
  - If you need to send emails to different domains, contact Postmark support to inquire about the approval process and any necessary steps to lift this restriction.

This error is designed to prevent misuse of the email service and ensure compliance with Postmark's policies during the account approval phase.

### Postmark Full Curl Payload Example
```
curl "https://api.postmarkapp.com/email" \
  -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H "X-Postmark-Server-Token: c4ffc3fe-d077-4480-95c2-629e203c919d" \
  -d '{
  "From": "nick@nzenitram.com",
  "To": "nzenitram@nzenitram.com",
  "Cc": "nick1@nzenitram.com",
  "Bcc": "support@nzenitram.com",
  "Subject": "Updating the subject to reflect the test",
  "Tag": "",
  "HtmlBody": "This is the body of the email.",
  "TextBody": "This is the body of the email.",
  "ReplyTo": "",
  "Metadata": null,
  "Headers": [
    {
      "Name": "X-Custom-Header-1",
      "Value": "Custom Value 1"
    },
    {
      "Name": "X-Custom-Header-2",
      "Value": "Custom Value 2"
    }
  ],
  "Attachments": [
    {
      "Name": "example.txt",
      "ContentType": "text/plain",
      "Content": "SGVsbG8gd29ybGQh"
    }
  ],
  "TrackOpens": true,
  "TrackLinks": "HtmlOnly",
  "MessageStream": "outbound"
}'
```

### Webhook Curl SendGrid Full

```
curl -X POST http://localhost:8888 -H "Content-Type: application/json" -d '[
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "processed",
    "category": ["cat facts"],
    "sg_event_id": "OU7xod-BfvaSsHt7yr90JQ==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0"
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "deferred",
    "category": ["cat facts"],
    "sg_event_id": "8kZc22_NI4O5txLANpCsew==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0",
    "response": "400 try again later",
    "attempt": "5"
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "delivered",
    "category": ["cat facts"],
    "sg_event_id": "B7MRBVJgwGSlUneYkWM4AA==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0",
    "response": "250 OK"
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "open",
    "category": ["cat facts"],
    "sg_event_id": "FSB8awrHyezzlKBNghxAIA==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0",
    "useragent": "Mozilla/4.0 (compatible; MSIE 6.1; Windows XP; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
    "ip": "255.255.255.255"
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "click",
    "category": ["cat facts"],
    "sg_event_id": "iPkmY5nE7yl4pcRpn4l88g==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0",
    "useragent": "Mozilla/4.0 (compatible; MSIE 6.1; Windows XP; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
    "ip": "255.255.255.255",
    "url": "http://www.sendgrid.com/"
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "bounce",
    "category": ["cat facts"],
    "sg_event_id": "Edt5m61-GwaKBAd9Yeithg==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0",
    "reason": "500 unknown recipient",
    "status": "5.0.0"
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "dropped",
    "category": ["cat facts"],
    "sg_event_id": "oLooansnxksZpNIrB2mGjw==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0",
    "reason": "Bounced Address",
    "status": "5.0.0"
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "spamreport",
    "category": ["cat facts"],
    "sg_event_id": "Hw-CL0uEOXrCijrpCrOFBA==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0"
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "unsubscribe",
    "category": ["cat facts"],
    "sg_event_id": "Z0n3N7FDyTCGL3xtb5c6Cw==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0"
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "group_unsubscribe",
    "category": ["cat facts"],
    "sg_event_id": "VCSBJ7b1bRBwoWs9t0KRIw==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0",
    "useragent": "Mozilla/4.0 (compatible; MSIE 6.1; Windows XP; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
    "ip": "255.255.255.255",
    "url": "http://www.sendgrid.com/",
    "asm_group_id": 10
  },
  {
    "email": "example@test.com",
    "timestamp": 1724813706,
    "smtp-id": "\u003c14c5d75ce93.dfd.64b469@ismtpd-555\u003e",
    "event": "group_resubscribe",
    "category": ["cat facts"],
    "sg_event_id": "9f_-AVK6xVQ8flu_AeFQiw==",
    "sg_message_id": "14c5d75ce93.dfd.64b469.filter0001.16648.5515E0B88.0",
    "useragent": "Mozilla/4.0 (compatible; MSIE 6.1; Windows XP; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
    "ip": "255.255.255.255",
    "url": "http://www.sendgrid.com/",
    "asm_group_id": 10
  }
]'
```


