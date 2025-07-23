package mail

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"message_handler/config"
	"regexp"
)

// MailDialer allows mocking gomail.Dialer
type MailDialer interface {
	DialAndSend(...*gomail.Message) error
}

// NewDialer constructs the real gomail dialer
func NewDialer(cfg *config.AppConfig) MailDialer {
	return gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
}

func sendMail(cfg *config.AppConfig, to, subject, body string, dialer MailDialer) error {
	if !validateEmail(to) {
		return fmt.Errorf("invalid email address: %s", to)
	}

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", cfg.SMTPUser)
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body)

	if err := dialer.DialAndSend(mailer); err != nil {
		return fmt.Errorf("failed to send email: %v. make sure to check the settings.env and recompile it if you changed the settings", err)
	}


	log.Println("Email sent successfully to: " + to + " Subject: `" + subject + "`")
	return nil
}

func validateEmail(email string) bool {
	regex := `^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`
	return regexp.MustCompile(regex).MatchString(email)
}
