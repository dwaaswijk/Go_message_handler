package mail

import (
	"errors"
	"gopkg.in/gomail.v2"
	"message_handler/config"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"os"
)

type MockDialer struct {
	EmailsSent []*gomail.Message
	SendError  error
}

func (m *MockDialer) DialAndSend(msg ...*gomail.Message) error {
	if m.SendError != nil {
		return m.SendError
	}
	m.EmailsSent = append(m.EmailsSent, msg...)
	return nil
}

func loadTestConfig(t *testing.T) *config.AppConfig {
	t.Helper()
	err := godotenv.Load("../settings.env")
	if err != nil {
		t.Fatalf("Failed to load settings.env: %v", err)
	}

	return &config.AppConfig{
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: func() int {
			port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
			return port
		}(),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
	}
}

func TestSendMail_Success(t *testing.T) {
	mockDialer := &MockDialer{}
	testConfig := loadTestConfig(t)

	err := sendMail(testConfig, "recipient@example.com", "Subject", "Body", mockDialer)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if len(mockDialer.EmailsSent) != 1 {
		t.Errorf("Expected 1 email to be sent, but got %d", len(mockDialer.EmailsSent))
	}
}

func TestSendMail_InvalidEmail(t *testing.T) {
	mockDialer := &MockDialer{}
	testConfig := loadTestConfig(t)

	err := sendMail(testConfig, "invalid-email", "Subject", "Body", mockDialer)
	if err == nil {
		t.Error("Expected error for invalid email, but got none")
	}

	expectedError := "invalid email address: invalid-email"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSendMail_SendError(t *testing.T) {
	mockDialer := &MockDialer{
		SendError: errors.New("SMTP connection failed"),
	}
	testConfig := loadTestConfig(t)

	err := sendMail(testConfig, "recipient@example.com", "Subject", "Body", mockDialer)
	if err == nil {
		t.Fatal("Expected error, but got none")
	}

	expectedError := "failed to send email: SMTP connection failed"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestHandleSendEmail_Success(t *testing.T) {
	mockDialer := &MockDialer{}
	testConfig := loadTestConfig(t)

	req := httptest.NewRequest(http.MethodPost, "/send-email", nil)
	req.Form = map[string][]string{
		"to":      {"test@example.com"},
		"subject": {"Subject"},
		"body":    {"Body"},
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := sendMail(testConfig, r.FormValue("to"), r.FormValue("subject"), r.FormValue("body"), mockDialer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Email sent successfully\n"))
	})

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	expectedResponse := "Email sent successfully\n"
	if rr.Body.String() != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, rr.Body.String())
	}
}

func TestHandleSendEmail_InvalidEmail(t *testing.T) {
	mockDialer := &MockDialer{}
	testConfig := loadTestConfig(t)

	req := httptest.NewRequest(http.MethodPost, "/send-email", nil)
	req.Form = map[string][]string{
		"to":      {"invalid-email"},
		"subject": {"Subject"},
		"body":    {"Body"},
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := sendMail(testConfig, r.FormValue("to"), r.FormValue("subject"), r.FormValue("body"), mockDialer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Email sent successfully\n"))
	})

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}

	expectedResponse := "invalid email address: invalid-email\n"
	if rr.Body.String() != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, rr.Body.String())
	}
}
