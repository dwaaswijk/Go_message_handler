package sms

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

// SendSMS sends an SMS using the provided serial port, recipient number, and message
func SendSMS(port io.ReadWriter, recipient, message string) error {
	log.Printf("Sending SMS to %s", maskPhone(recipient))

	// Set SMS to text mode
	cmd := "AT+CMGF=1" + "\r"
	if _, err := port.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to set text mode: %w", err)
	}

	time.Sleep(1 * time.Second)

	// Send recipient number
	cmd = fmt.Sprintf(`AT+CMGS="%s"`, recipient) + "\r"
	if _, err := port.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to send phone number: %w", err)
	}

	time.Sleep(1 * time.Second)

	// Send message and end with Ctrl+Z
	if _, err := port.Write([]byte(message + string(26))); err != nil { // 26 is Ctrl+Z
		return fmt.Errorf("failed to send message: %w", err)
	}

	response := make([]byte, 1024)
	n, _ := port.Read(response)
	if !strings.Contains(string(response[:n]), "OK") {
		return fmt.Errorf("failed to send SMS, modem response: %s", string(response[:n]))
	}

	return nil
}

// ValidatePhone checks if the phone number is in a valid E.164 format
func ValidatePhone(phone string) bool {
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}
	return strings.HasPrefix(phone, "+")
}

// maskPhone obfuscates the phone number for logging
func maskPhone(phone string) string {
	if len(phone) > 4 {
		return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
	}
	return "****"
}
