package service

import (
	"encoding/json"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/mqtt"
)

// ReplayService publishes a stored IR code back to the device.
type ReplayService struct {
	client       mqtt.MQTTClient
	publishTopic string
}

// NewReplayService constructs a ReplayService for the given device publish topic.
func NewReplayService(client mqtt.MQTTClient, publishTopic string) *ReplayService {
	return &ReplayService{client: client, publishTopic: publishTopic}
}

// Replay publishes {"ir_code_to_send": base64Payload} to the device.
func (s *ReplayService) Replay(base64Payload string) error {
	payload, _ := json.Marshal(map[string]string{"ir_code_to_send": base64Payload})
	return s.client.Publish(s.publishTopic, payload)
}
