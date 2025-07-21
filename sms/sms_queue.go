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

// SMSQueue handles SMS sending in a queue
type SMSQueue struct {
	queue  chan *SMS
	stopCh chan struct{}
	wg     sync.WaitGroup
	send   func(*SMS)
}

// NewSMSQueue initializes an SMS queue with the given capacity
func NewSMSQueue(capacity int) *SMSQueue {
	return &SMSQueue{
		queue:  make(chan *SMS, capacity),
		stopCh: make(chan struct{}),
	}
}

// SetSender defines the sending function for SMSQueue
func (q *SMSQueue) SetSender(sendFunc func(*SMS)) {
	q.send = sendFunc
}

// Send queues an SMS message for sending
func (q *SMSQueue) Send(sms *SMS) {
	q.queue <- sms
}

// Start begins the SMSQueue processing
func (q *SMSQueue) Start() {
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		for {
			select {
			case sms := <-q.queue:
				if q.send != nil {
					q.send(sms)
				} else {
					log.Println("No sender configured for SMSQueue")
				}
			case <-q.stopCh:
				return
			}
		}
	}()
}

// Stop shuts down the SMSQueue
func (q *SMSQueue) Stop() {
	close(q.stopCh)
	q.wg.Wait()
}
