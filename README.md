## Email Sending Endpoint JSON Structure

This section describes the JSON structure used to send an email through the email sending endpoint. The JSON object contains various fields that specify the sender, recipients, email content, and additional options.

### JSON Fields

- **`from`**: A string representing the sender's email address and name. It follows the format `"Name <email@example.com>"`.

- **`to`**: An array of strings, each representing a recipient's email address and name. Each entry follows the format `"Name <email@example.com>"`.

- **`cc`**: An array of strings, each representing a CC (carbon copy) recipient's email address. This field is optional.

- **`bcc`**: An array of strings, each representing a BCC (blind carbon copy) recipient's email address. This field is optional.

- **`subject`**: A string representing the subject line of the email.

- **`body`**: A string containing the main content of the email.

- **`attachments`**: An array of objects, each representing an attachment. Each attachment object contains:
  - **`name`**: A string specifying the file name of the attachment.
  - **`contenttype`**: A string specifying the MIME type of the attachment.
  - **`content`**: A base64-encoded string representing the content of the attachment.

- **`headers`**: An object containing custom email headers. Each key-value pair represents a header name and its corresponding value.

- **`data`**: An object containing additional email sending options:
  - **`TrackOpens`**: A boolean indicating whether to track email opens.
  - **`TrackLinks`**: A string specifying link tracking options. Possible values include `"HtmlOnly"`.
  - **`MessageStream`**: A string specifying the message stream to use, such as `"outbound"`.

This JSON structure allows for detailed customization of the email sending process, including specifying recipients, adding attachments, and setting tracking options.

## Example CURL Requests for Testing 

### Sample JSON Structure for Email Sends
```
'{
  "from": "Twitter Zen <twitter1@nzenitram.com>",
  "to": [
    "\"Nick Martinez, Jr.\" <nick@nzenitram.com>",
    "Mick Nartinez <nzenitram@nzenitram.com>"
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
  },
  "data": {
    "TrackOpens": true,
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
  }
}'
```


### Curl with Friendly From, Headers and Attchment
```
curl -X POST http://localhost:8888 -H "Content-Type: application/json" -d '{
  "from": "Twitter Zen <twitter1@nzenitram.com>",
  "to": [
    "\"Nick Martinez, Jr.\" <nick@nzenitram.com>",
    "Mick Nartinez <nzenitram@nzenitram.com>"
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
  },
  "data": {
    "TrackOpens": true,
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
  }
}'
```

### Curl without Friendly From
```
curl -X POST http://localhost:8888 -H "Content-Type: application/json" -d '{
  "from": "<twitter1@nzenitram.com>",
  "to": [
    "<nick@nzenitram.com>",
    "<nzenitram@nzenitram.com>"
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
  },
  "data": {
    "TrackOpens": true,
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
  }
}'
```

### Curl without Friendly From and brackets
```
curl -X POST http://localhost:8888 -H "Content-Type: application/json" -d '{
  "from": "twitter1@nzenitram.com",
  "to": [
    "nick@nzenitram.com",
    "nzenitram@nzenitram.com"
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
  },
  "data": {
    "TrackOpens": true,
    "TrackLinks": "HtmlOnly",
    "MessageStream": "outbound"
  }
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