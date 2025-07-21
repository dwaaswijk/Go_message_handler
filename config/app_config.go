package config

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
