package mail

import (
	"log"
	"message_handler/config"
	"net/http"
)

func HandleSendEmail(w http.ResponseWriter, r *http.Request, cfg *config.AppConfig) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	to := r.FormValue("to")
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	if to == "" || subject == "" || body == "" {
		http.Error(w, "Missing email fields", http.StatusBadRequest)
		return
	}
	if !validateEmail(to) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	dialer := NewDialer(cfg)
	if err := sendMail(cfg, to, subject, body, dialer); err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Email sent successfully\n"))
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
