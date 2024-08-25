### Curl with Friendly From, Headers and Attchment
```
curl -X POST http://localhost:8888 -H "Content-Type: application/json" -d '{
  "from": {
    "email": "nick@nzenitram.com",
    "name": "Twitter Zen"
  },
  "to": [
    {
      "email": "nick@nzenitram.com",
      "name": "Nick Zenitram"
    },
    {
      "email": "nicholas@nzenitram.com",
      "name": "Recipient Two"
    }
  ],
  "cc": ["nick1@nzenitram.com"],
  "bcc": ["support@clickaleague.com"],
  "subject": "Updating the subject to reflect the test",
  "body": "This is the body of the email.",
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
  }
}'
```

### Curl with Friendly From
```
curl -X POST http://localhost:8888 -H "Content-Type: application/json" -d '{
  "from": {
    "email": "twitter1@nzenitram.com",
    "name": "Twitter Zen"
  },
  "to": [
    {
      "email": "nick@nzenitram.com",
      "name": "Nick Zenitram"
    },
    {
      "email": "nicholas@nzenitram.com",
      "name": "Recipient Two"
    }
  ],
  "cc": ["nick1@nzenitram.com"],
  "bcc": ["support@clickaleague.com"],
  "subject": "Updating the subject to reflect the test",
  "body": "This is the body of the email."
}'
```

### Curl without Friendly From
```
curl -X POST http://localhost:8888 -H "Content-Type: application/json" -d '{
  "from": "sender@example.com",
  "to": ["nick@nzenitram.com", "recipient2@example.com"],
  "cc": ["cc1@example.com"],
  "bcc": ["bcc1@example.com"],
  "subject": "Example Subject",
  "body": "This is the body of the email.",
  "attachments": [
    {
      "filename": "example.txt",
      "content_type": "text/plain",
      "content": "SGVsbG8gd29ybGQh"
    }
  ]
}'
```

### Postmark Full Curl Payload Example
```
curl "https://api.postmarkapp.com/email" \
  -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H "X-Postmark-Server-Token: server token" \
  -d '{
  "From": "Twitter Zen <twitter1@nzenitram.com>",
  "To": "\"Nick Martinez, Jr.\" <nick@nzenitram.com>",
  "Cc": "nick1@nzenitram.com",
  "Bcc": "support@clickaleague.com",
  "Subject": "This is a test PostMark Email",
  "Tag": "Invitation",
  "HtmlBody": "<b>Hello</b>",
  "TextBody": "Hello",
  "ReplyTo": "twitter1@nzenitram.com",
  "Metadata": {
      "Color":"blue",
      "Client-Id":"12345"
  },
  "Headers": [
    {
      "Name": "CUSTOM-HEADER",
      "Value": "value"
    }
  ],
  "Attachments": [
    {
      "Name": "readme.txt",
      "Content": "dGVzdCBjb250ZW50",
      "ContentType": "text/plain"
    },
    {
      "Name": "report.pdf",
      "Content": "dGVzdCBjb250ZW50",
      "ContentType": "application/octet-stream"
    }
  ],
  "AddtionalData":[
  "TrackOpens": true,
  "TrackLinks": "HtmlOnly",
  "MessageStream": "outbound"
  ]
}
```