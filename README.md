# Messaging Service Application

This project provides a messaging service that supports sending SMS through a modem and sending Emails via an SMTP server. The application also includes built-in rate limiting to prevent abuse and customizable configuration to suit different setups.

---

## Features

1. **Send SMS**: Use an AT-compatible modem to send SMS.
2. **Send Emails**: Send emails using SMTP.
3. **Rate Limiting**: Limits requests per second per IP to avoid abuse.
4. **Secure Endpoints**: API key validation for protected endpoints.
5. **Environment Configuration**: All settings are configurable through environment variables or a `.env` file (`settings.env`).

---

## Getting Started

Follow these steps to install and run the application:

### Prerequisites

- Install **Go (v1.24)**.
- Have an AT-command compatible modem (e.g., a GSM modem).
- SMTP account credentials (such as Gmail or other SMTP services).
- API key for securing endpoints.

### Installation

1. **Unzip the codebase**:
   Download and unzip the project folder.

2. **Navigate to the project directory**:
   ```bash
   cd <project-folder>
   ```

3. **Install dependencies**:
   Run the following command to install the required dependencies.
   ```bash
   go mod tidy
   ```

4. **Set up environment variables**:
   Create and configure a `settings.env` file in the project's root directory, or ensure the environment variables are correctly set up in your system. See the **Configuration** section below for a detailed explanation.


5. **build project**:
   Run the following command to build the executable binary
   ```bash
   go build
   ```
---

## Running Unit Tests

This project includes unit tests to ensure the reliability of its functionality.

### How to run the tests

1. Ensure all dependencies are installed:
   ```bash
   go mod tidy
   ```

2. Run the unit tests using the `go test` command:
   ```bash
   go test ./...
   ```

   This will run all test files in the project.

### Test Coverage
To see the test coverage, you can use the `-cover` flag:

---

## Configuration

The application works with environment variables. You can use a `.env` file called `settings.env` (example included) for easy setup.

### Example `settings.env`
```
# API Key for requests
API_KEY=PUTYOURAPIKEYHERE

# Server configuration
SERVER_PORT=8080

# Rate limiter configuration
RATE_LIMIT=2          # 2 requests per second
BURST_LIMIT=10        # 10 requests burst capacity

# SMS configuration
MAX_QUEUE_SIZE=5
SMS_PROVIDER=twilio   # Options: "hardware" or "twilio"

# For hardware
# SERIAL_BAUD=9600

# For Twilio
TWILIO_SID=TWILIOSID
TWILIO_AUTH_TOKEN=AUTHTOKEN
TWILIO_PHONE=TWILIOPHONE

# SMTP server configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=YOUREMAILHERE
SMTP_PASS=YOURAPPPASWORDHERE
```

## Example CURL Requests

Here are some example `curl` requests to demonstrate how to use the API endpoints.

### 1. Send an SMS
This endpoint allows you to send an SMS using the configured modem.

**Endpoint:**
POST /api/sms

```bash
curl -v -X POST http://localhost:8080/send-sms \
  -H "Authorization: PUTYOURAPIKEYHERE" \
  --data-urlencode "phone=+1234567890" \
  --data-urlencode "message=Testing + symbol"
```

**Parameters:**
- `recipient`: The phone number of the message receiver (in international format).
- `message`: The message content.

---

### 2. Send an Email
This endpoint allows you to send an email using the configured SMTP server.

**Endpoint:** POST /api/email

```bash
curl -X POST http://localhost:8080/send-email \
  -H "Authorization: PUTYOURAPIKEYHERE" \
  --form "to=example@example.com" \
  --form "subject=HTML Test Email" \
  --form-string "body=<h1>Hello!</h1><p>This is a <strong>test</strong> email with <a href='https://example.com'>HTML content</a>.</p>"
```

**Parameters:**
- `to`: The email address of the recipient.
- `subject`: The subject of the email.
- `body`: The email body/content.

---

