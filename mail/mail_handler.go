package mail

import (
	"log"
	"message_handler/config"
	"net/http"
	"strings"
)

func HandleSendEmail(w http.ResponseWriter, r *http.Request, cfg *config.AppConfig) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	to := r.FormValue("to")
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	// Check for missing required fields
	if strings.TrimSpace(to) == "" || strings.TrimSpace(subject) == "" || strings.TrimSpace(body) == "" {
		http.Error(w, "Missing required email fields", http.StatusBadRequest)
		return
	}

	// Validate the email address format
	if !validateEmail(to) {
		http.Error(w, "Invalid email address format", http.StatusBadRequest)
		return
	}

	// Create a new dialer and attempt to send the email
	dialer := NewDialer(cfg)
	err := sendMail(cfg, to, subject, body, dialer)

	// Handle send-mail errors appropriately
	if err != nil {
		var errMessage string
		if strings.Contains(err.Error(), "invalid email address") {
			// Specific error for email validation issues
			errMessage = "Invalid recipient email address"
			http.Error(w, errMessage, http.StatusBadRequest)
		} else {
			// General failure during email sending
			errMessage = "Failed to send email due to an internal server issue"
			http.Error(w, errMessage, http.StatusInternalServerError)
		}
		log.Printf("Error while sending email: %v", err)
		return
	}

	// Successful response
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Email sent successfully\n"))
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}