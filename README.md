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

## Configuration

The application works with environment variables. You can use a `.env` file called `settings.env` (example included) for easy setup.

### Example `settings.env`
```
# API Key for requests
SMS_API_KEY=PUTYOURAPIKEYHERE

# Server configuration
SERVER_PORT=8080

# Rate limiter configuration
RATE_LIMIT=2          # 2 requests per second
BURST_LIMIT=10        # 10 requests burst capacity

# Maximum SMS queue size
MAX_QUEUE_SIZE=100 

# SMTP server configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-email-password
```

## Example CURL Requests

Here are some example `curl` requests to demonstrate how to use the API endpoints.

### 1. Send an SMS
This endpoint allows you to send an SMS using the configured modem.

**Endpoint:**
POST /api/sms

```bash
curl -X POST [http://localhost:8080/api/sms](http://localhost:8080/api/sms)
-H "Content-Type: application/json"
-H "Authorization: Bearer <YOUR_API_KEY>"
-d '{ "recipient": "+1234567890", "message": "Hello! This is a test SMS." }'

```

**Parameters:**
- `recipient`: The phone number of the message receiver (in international format).
- `message`: The message content.

---

### 2. Send an Email
This endpoint allows you to send an email using the configured SMTP server.

**Endpoint:** POST /api/email

```bash
 curl -X POST [http://localhost:8080/api/email](http://localhost:8080/api/email)
-H "Content-Type: application/json"
-H "Authorization: Bearer <YOUR_API_KEY>"
-d '{ "to": "example@domain.com", "subject": "Test Email", "body": "This is a test email sent from the Messaging Service" }'
```
**Parameters:**
- `to`: The email address of the recipient.
- `subject`: The subject of the email.
- `body`: The email body/content.

---

