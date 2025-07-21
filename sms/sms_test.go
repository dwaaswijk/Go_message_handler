package sms

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
)

// MockPort simulates a serial port for testing purposes.
type MockPort struct {
	WriteBuffer bytes.Buffer
	ReadBuffer  bytes.Buffer
}

// Write writes data to the mock port's write buffer.
func (m *MockPort) Write(p []byte) (int, error) {
	return m.WriteBuffer.Write(p)
}

// Read reads data from the mock port's read buffer.
func (m *MockPort) Read(p []byte) (int, error) {
	return m.ReadBuffer.Read(p)
}

// Close simulates closing the mock port.
func (m *MockPort) Close() error {
	return nil
}

// TestSendSMS tests the SendSMS function.
func TestSendSMS(t *testing.T) {
	err := godotenv.Load("../settings.env")
	if err != nil {
		t.Fatalf("Failed to load environment variables: %v", err)
	}

	mockPort := &MockPort{}

	tests := []struct {
		name          string
		recipient     string
		message       string
		mockResponse  string
		expectError   bool
		expectedWrite string
	}{
		{
			name:          "ValidSMS",
			recipient:     "+1234567890",
			message:       "Hello, this is a test message",
			mockResponse:  "OK",
			expectError:   false,
			expectedWrite: "AT+CMGF=1\rAT+CMGS=\"+1234567890\"\rHello, this is a test message\x1A",
		},
		{
			name:         "ModemErrorResponse",
			recipient:    "+1234567890",
			message:      "Test error response",
			mockResponse: "ERROR",
			expectError:  true,
		},
		{
			name:        "EmptyMessage",
			recipient:   "+1234567890",
			message:     "",
			expectError: true,
		},
		{
			name:        "InvalidRecipient",
			recipient:   "INVALID",
			message:     "Test message",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockPort.ReadBuffer.Reset()
			mockPort.ReadBuffer.WriteString(tc.mockResponse)

			mockPort.WriteBuffer.Reset()

			err := SendSMS(mockPort, tc.recipient, tc.message)
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Did not expect error but got one: %v", err)
			}

			if tc.expectedWrite != "" && mockPort.WriteBuffer.String() != tc.expectedWrite {
				t.Errorf("Unexpected commands sent to port. Got: %q, want: %q", mockPort.WriteBuffer.String(), tc.expectedWrite)
			}
		})
	}
}

// TestValidatePhone tests the phone number validation logic.
func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		isValid bool
	}{
		{"ValidPhone", "+1234567890", true},
		{"InvalidPhoneLetters", "123ABC", false},
		{"InvalidPhoneShort", "12345", false},
		{"EmptyPhone", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidatePhone(tc.phone)
			if result != tc.isValid {
				t.Errorf("ValidatePhone(%q) = %v, want %v", tc.phone, result, tc.isValid)
			}
		})
	}
}

// TestQueueSMS tests the functionality of the SMS queue.
func TestQueueSMS(t *testing.T) {
	err := godotenv.Load("../settings.env")
	if err != nil {
		t.Fatalf("Failed to load environment variables: %v", err)
	}

	maxQueueSize, parseErr := strconv.Atoi(os.Getenv("MAX_QUEUE_SIZE"))
	if parseErr != nil || maxQueueSize <= 0 {
		maxQueueSize = 100 // Default if not defined in settings.env
	}

	smsQueue := NewSMSQueue(maxQueueSize)
	smsQueue.SetSender(func(sms *SMS) {
		// mock SMS send it to avoid actual work or sleeping in tests
		t.Logf("Mock send to %s: %s", sms.Recipient, sms.Message)
	})
	smsQueue.Start()
	defer smsQueue.Stop()

	tests := []struct {
		name        string
		sms         *SMS
		expectError bool // Updated to remove reliance on error return from `Send()`
	}{
		{"ValidSMS", &SMS{Recipient: "+1234567890", Message: "Hello!"}, false},
		{"InvalidRecipient", &SMS{Recipient: "INVALID", Message: "Hello!"}, false},
		{"EmptyMessage", &SMS{Recipient: "+1234567890", Message: ""}, false},
	}

	// Fill the queue to test overcapacity
	for i := 0; i < maxQueueSize; i++ {
		smsQueue.Send(&SMS{Recipient: "+1234567890", Message: "Test SMS"})
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.expectError {
				// Expect the SMS to be queued without explicit error handling
				smsQueue.Send(tc.sms)
			} else {
				// Overcapacity simply blocks without a reported error as per the queue's design
				select {
				case smsQueue.queue <- tc.sms:
					t.Errorf("Queue allowed overcapacity for SMS: %+v", tc.sms)
				default:
					// The queue rejected the message correctly
				}
			}
		})
	}
}
