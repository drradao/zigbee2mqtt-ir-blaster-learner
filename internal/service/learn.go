// Package service contains the business logic for learning and replaying IR
// codes. Services depend only on the mqtt.MQTTClient interface so they can
// be tested without a real broker.
package service

import (
	"sync"
	"time"

	"github.com/drradao/zigbee2mqtt-ir-blaster-learner/internal/mqtt"
)

// learnCommandPayload is the static JSON payload sent to trigger IR learning.
var learnCommandPayload = []byte(`{"learn_ir_code":true}`)

// LearnService sends "learn mode" commands to the device.
type LearnService struct {
	client       mqtt.MQTTClient
	publishTopic string
	mu           sync.Mutex
	lastSent     time.Time
}

// NewLearnService constructs a LearnService for the given device publish topic.
func NewLearnService(client mqtt.MQTTClient, publishTopic string) *LearnService {
	return &LearnService{
		client:       client,
		publishTopic: publishTopic,
	}
}

// SendLearnCommand publishes {"learn_ir_code": true} to the device.
func (s *LearnService) SendLearnCommand() error {
	s.mu.Lock()
	s.lastSent = time.Now()
	s.mu.Unlock()
	return s.client.Publish(s.publishTopic, learnCommandPayload)
}

// LastSent returns the time the last learn command was sent.
func (s *LearnService) LastSent() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSent
}
