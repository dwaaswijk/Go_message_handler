package sms

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// SendSMSviaTwilio sends SMS using Twilio's API
func SendSMSviaTwilio(accountSID, authToken, twilioNumber string, sms *SMS) error {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	params := &openapi.CreateMessageParams{}
	params.SetTo(sms.Recipient)
	params.SetFrom(twilioNumber)
	params.SetBody(sms.Message)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		return err
	}

	log.Printf("Twilio SMS sent: SID=%s Status=%s", *resp.Sid, *resp.Status)
	return nil
}

// SendSMSviaHardware sends an SMS using a hardware modem over serial
func SendSMSviaHardware(port io.ReadWriter, recipient, message string) error {
	log.Printf("Sending SMS to %s", maskPhone(recipient))

	// Set SMS to text mode
	cmd := "AT+CMGF=1\r"
	if _, err := port.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to set text mode: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Set recipient
	cmd = fmt.Sprintf(`AT+CMGS="%s"`+"\r", recipient)
	if _, err := port.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to send phone number: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Send message, followed by Ctrl+Z
	if _, err := port.Write([]byte(message + string(rune(26)))); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Read modem response
	response := make([]byte, 1024)
	n, _ := port.Read(response)
	if !strings.Contains(string(response[:n]), "OK") {
		return fmt.Errorf("failed to send SMS, modem response: %s", string(response[:n]))
	}

	return nil
}
