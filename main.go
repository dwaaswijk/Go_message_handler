package main

import (
	"fmt"
	"io"
	"log"
	"message_handler/sms"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

var (
	portMutex sync.Mutex // Ensures thread safety for the modem port
	apiKey    string     // API key for securing the endpoints
	smsQueue  *sms.SMSQueue
)

type AppConfig struct {
	ServerPort   string
	RateLimit    float64 // Requests per second
	BurstLimit   int     // Burst requests allowed
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
	DevicePath   string // Path to the serial device
	MaxQueueSize int    // Maximum SMS queue size
}

// Update loadConfig function to include MaxQueueSize
func loadConfig() (*AppConfig, error) {
	err := godotenv.Load("settings.env")
	if err != nil {
		log.Println("Warning: Could not load settings.env. Falling back to system environment variables")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "5643"
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
		maxQueueSize = 100 // Default value
	}

	return &AppConfig{
		ServerPort:   serverPort,
		RateLimit:    rateLimit,
		BurstLimit:   burstLimit,
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     smtpPort,
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPass:     os.Getenv("SMTP_PASS"),
		DevicePath:   os.Getenv("DEVICE_PATH"),
		MaxQueueSize: maxQueueSize,
	}, nil
}

// openSerialPort opens the serial port at the specified path
func openSerialPort(devicePath string) (io.ReadWriteCloser, error) {
	// Replace with actual serial port opening logic
	// Here we use a mock implementation for demonstration
	return os.OpenFile(devicePath, os.O_RDWR, 0666)
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	apiKey = os.Getenv("SMS_API_KEY")
	if apiKey == "" {
		log.Println("Warning: SMS_API_KEY environment variable is not set. SMS sending will be disabled.")
	}

	var serialPort io.ReadWriteCloser
	serialPortError := ""
	if apiKey != "" {
		serialPort, err = openSerialPort(config.DevicePath)
		if err != nil {
			serialPortError = fmt.Sprintf("Failed to open serial port: %v", err)
			log.Println(serialPortError)
		} else {
			defer serialPort.Close()
		}
	}

	// Initialize the SMS queue with the configured max queue size
	smsQueue = sms.NewSMSQueue(config.MaxQueueSize)
	defer smsQueue.Stop()

	// Set real sending function instead of dummy
	if serialPort != nil {
		smsQueue.SetSender(func(s *sms.SMS) {
			portMutex.Lock()
			defer portMutex.Unlock()
			err := sms.SendSMS(serialPort, s.Recipient, s.Message)
			if err != nil {
				log.Printf("SMS send error: %v", err)
			}
		})
	}

	// Initialize rate limiter
	rl := NewRateLimiter(rate.Limit(config.RateLimit), config.BurstLimit)

	// Routes
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	http.Handle("/send-sms", rl.LimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serialPort == nil || serialPortError != "" {
			http.Error(w, "SMS service is not configured or unavailable", http.StatusNotImplemented)
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

	log.Printf("Server is listening on port %s...", config.ServerPort)
	if err := http.ListenAndServe(":"+config.ServerPort, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// authenticate checks whether the request contains the correct API key
func authenticate(r *http.Request) bool {
	clientAPIKey := r.Header.Get("Authorization")
	return clientAPIKey == apiKey
}
