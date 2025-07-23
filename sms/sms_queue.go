package sms

import (
	"log"
	"sync"

)

// SMS represents a single SMS message
type SMS struct {
	Recipient string
	Message   string
}

// SMSQueue handles SMS sending in a queue with multiple sender options
type SMSQueue struct {
	queue       chan *SMS
	stopCh      chan struct{}
	wg          sync.WaitGroup
	hardwareSend func(*SMS) error // Hardware-based sender
	twilioSend   func(*SMS) error // Twilio-based sender
	provider     string           // Selected provider ("hardware" or "twilio")
}

// SetProvider sets the preferred SMS provider
func (q *SMSQueue) SetProvider(provider string) {
	q.provider = provider
}

// SetHardwareSender configures the hardware SMS sender
func (q *SMSQueue) SetHardwareSender(sendFunc func(*SMS) error) {
	q.hardwareSend = sendFunc
}

// SetTwilioSender configures the Twilio SMS sender
func (q *SMSQueue) SetTwilioSender(sendFunc func(*SMS) error) {
	q.twilioSend = sendFunc
}

// Send queues an SMS message for sending
func (q *SMSQueue) Send(sms *SMS) {
	q.queue <- sms
}

// Start begins processing the SMS queue
func (q *SMSQueue) Start() {
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		for {
			select {
			case sms := <-q.queue:
				var err error
				if q.provider == "twilio" && q.twilioSend != nil {
					err = q.twilioSend(sms)
				} else if q.provider == "hardware" && q.hardwareSend != nil {
					err = q.hardwareSend(sms)
				} else {
					log.Println("No sender configured or provider not set correctly")
				}
				if err != nil {
					log.Printf("Failed to send SMS to %s: %v", sms.Recipient, err)
				}
			case <-q.stopCh:
				return
			}
		}
	}()
}

func NewSMSQueue(bufferSize int) *SMSQueue {
	return &SMSQueue{
		queue:   make(chan *SMS, bufferSize),
		stopCh:  make(chan struct{}),
	}
}

func (q *SMSQueue) Stop() {
	close(q.stopCh)
	q.wg.Wait()
}
