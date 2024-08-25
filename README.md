## Email Sending Endpoint JSON Structure

This section describes the JSON structure used to send an email through the email sending endpoint. The JSON object contains various fields that specify the sender, recipients, email content, additional options, and credentials for API authentication.

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
    "PostmarkAPIURL": "https://api.postmarkapp.com/email",
    "PostmarkWeight": "50"
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
    "PostmarkAPIURL": "https://api.postmarkapp.com/email",
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
    "PostmarkAPIURL": "https://api.postmarkapp.com/email",
    "PostmarkWeight": "50"
  }
}'
```

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