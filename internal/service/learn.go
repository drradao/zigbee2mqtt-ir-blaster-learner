// Package service contains the business logic for learning and replaying IR
// codes. Services depend only on the mqtt.MQTTClient interface so they can
// be tested without a real broker.
package service

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/mqtt"
)

// LearnService sends "learn mode" commands to the device and tracks lock mode.
type LearnService struct {
	client       mqtt.MQTTClient
	publishTopic string
	mu           sync.Mutex
	lockMode     bool
	lastSent     time.Time
	debounce     time.Duration
}

// NewLearnService constructs a LearnService for the given device publish topic.
func NewLearnService(client mqtt.MQTTClient, publishTopic string) *LearnService {
	return &LearnService{
		client:       client,
		publishTopic: publishTopic,
		debounce:     200 * time.Millisecond,
	}
}

// SendLearnCommand publishes {"learn_ir_code": true} to the device.
func (s *LearnService) SendLearnCommand() error {
	payload, _ := json.Marshal(map[string]interface{}{"learn_ir_code": true})
	s.mu.Lock()
	s.lastSent = time.Now()
	s.mu.Unlock()
	return s.client.Publish(s.publishTopic, payload)
}

// ToggleLockMode flips lock mode and returns the new state. In lock mode the
// TUI automatically re-sends the learn command after each captured code.
func (s *LearnService) ToggleLockMode() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lockMode = !s.lockMode
	return s.lockMode
}

// IsLockMode reports whether lock mode is currently active.
func (s *LearnService) IsLockMode() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lockMode
}

// LastSent returns the time the last learn command was sent.
func (s *LearnService) LastSent() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSent
}

// MaybeSendLearn sends the learn command if lock mode is active and the
// debounce window has elapsed.
func (s *LearnService) MaybeSendLearn() error {
	s.mu.Lock()
	if !s.lockMode || time.Since(s.lastSent) < s.debounce {
		s.mu.Unlock()
		return nil
	}
	s.lastSent = time.Now()
	s.mu.Unlock()
	return s.SendLearnCommand()
}
