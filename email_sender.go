package main

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"net/http"
	"regexp"
)

func sendMail(config *AppConfig, to, subject, body string) error {
	if !validateEmail(to) {
		return fmt.Errorf("invalid email address: %s", to)
	}

	// Prepare the email message
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", config.SMTPUser)
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/plain", body)

	// Connect to the mail server
	dialer := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPUser, config.SMTPPass)

	// Send the email
	if err := dialer.DialAndSend(mailer); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Println("Email sent successfully")
	return nil
}

func validateEmail(email string) bool {
	regex := `^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$` // E.164 format: +[country_code][number]
	return regexp.MustCompile(regex).MatchString(email)
}

func handleSendEmail(w http.ResponseWriter, r *http.Request, config *AppConfig) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	to := r.FormValue("to")
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	// Validate fields and email format
	if to == "" || subject == "" || body == "" {
		http.Error(w, "Missing email fields", http.StatusBadRequest)
		return
	}
	if !validateEmail(to) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	// Call sendMail and handle errors cleanly
	err := sendMail(config, to, subject, body)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Email sent successfully\n"))
	if err != nil {
		log.Printf("Failed to write response for email endpoint: %v", err)
	}
}
