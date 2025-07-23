package main

import (
	"fmt"
	"io"
	"log"
	"message_handler/config"
	"message_handler/mail"
	"message_handler/sms"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/tarm/serial"
	"golang.org/x/time/rate"
)

var (
	portMutex   sync.Mutex
	apiKey      string
	smsQueue    *sms.SMSQueue
	smsProvider string // Declare smsProvider as a package-level variable
)

func loadConfig() (*config.AppConfig, error) {
	err := godotenv.Load("settings.env")
	if err != nil {
		log.Println("Warning: Could not load settings.env. Falling back to system environment variables")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "5643"
	}

	// Assign value to the global smsProvider variable
	smsProvider = strings.ToLower(os.Getenv("SMS_PROVIDER"))
	if smsProvider != "hardware" && smsProvider != "twilio" {
		smsProvider = "hardware" // default
	}

	rateLimit, err := strconv.ParseFloat(os.Getenv("RATE_LIMIT"), 64)
	if err != nil || rateLimit <= 0 {
		rateLimit = 1
	}

	burstLimit, err := strconv.Atoi(os.Getenv("BURST_LIMIT"))
	if err != nil || burstLimit <= 0 {
		burstLimit = 5
	}

	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil || smtpPort <= 0 {
		smtpPort = 587
	}

	maxQueueSize, err := strconv.Atoi(os.Getenv("MAX_QUEUE_SIZE"))
	if err != nil || maxQueueSize <= 0 {
		maxQueueSize = 100
	}

	serialBaud := 115200
	if val, err := strconv.Atoi(os.Getenv("SERIAL_BAUD")); err == nil && val > 0 {
		serialBaud = val
	}

	return &config.AppConfig{
		ServerPort:   serverPort,
		RateLimit:    rateLimit,
		BurstLimit:   burstLimit,
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     smtpPort,
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPass:     os.Getenv("SMTP_PASS"),
		DevicePath:   os.Getenv("DEVICE_PATH"),
		MaxQueueSize: maxQueueSize,
		SerialBaud:   serialBaud,
	}, nil
}

func openSerialPort(devicePath string, baudRate int) (io.ReadWriteCloser, error) {
	config := &serial.Config{
		Name: devicePath,
		Baud: baudRate,
	}
	return serial.OpenPort(config)
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		log.Println("Warning: API_KEY environment variable is not set.")
	}

	baudRate := cfg.SerialBaud

	var serialPort io.ReadWriteCloser
	serialPortError := ""
	// smsProvider is now accessible here
	if smsProvider == "hardware" {
		if apiKey != "" {
			serialPort, err = openSerialPort(cfg.DevicePath, baudRate)
			if err != nil {
				serialPortError = fmt.Sprintf("Failed to open serial port: %v", err)
				log.Println(serialPortError)
			} else {
				defer serialPort.Close()
			}
		}
	}

	smsQueue = sms.NewSMSQueue(cfg.MaxQueueSize)
	smsQueue.SetProvider("hardware")
	smsQueue.Start()
	defer smsQueue.Stop()

	// Set hardware sender
	if serialPort != nil {
		smsQueue.SetHardwareSender(func(s *sms.SMS) error {
			portMutex.Lock()
			defer portMutex.Unlock()
			return sms.SendSMSviaHardware(serialPort, s.Recipient, s.Message)
		})
	}

	// Twilio setup
	twilioSID := os.Getenv("TWILIO_SID")
	twilioAuth := os.Getenv("TWILIO_AUTH_TOKEN")
	twilioNumber := os.Getenv("TWILIO_PHONE")
	if twilioSID != "" && twilioAuth != "" && twilioNumber != "" {
		smsQueue.SetProvider("twilio")
		smsQueue.SetTwilioSender(func(s *sms.SMS) error {
			return sms.SendSMSviaTwilio(twilioSID, twilioAuth, twilioNumber, s)
		})
	}

	rl := NewRateLimiter(rate.Limit(cfg.RateLimit), cfg.BurstLimit)

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	http.Handle("/send-sms", rl.LimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if smsQueue == nil {
			http.Error(w, "SMS service not initialized", http.StatusInternalServerError)
			return
		}
		if !authenticate(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		phone := r.FormValue("phone")
		message := r.FormValue("message")

		if !sms.ValidatePhone(phone) {
			http.Error(w, "Invalid phone number format", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(message) == "" {
			http.Error(w, "Message cannot be empty", http.StatusBadRequest)
			return
		}

		smsQueue.Send(&sms.SMS{Recipient: phone, Message: message})
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("SMS queued successfully\n"))
	})))

	http.Handle("/send-email", rl.LimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !authenticate(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		mail.HandleSendEmail(w, r, cfg)
	})))

	log.Printf("Server is listening on port %s...", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func authenticate(r *http.Request) bool {
	clientAPIKey := r.Header.Get("Authorization")
	return clientAPIKey == apiKey
}