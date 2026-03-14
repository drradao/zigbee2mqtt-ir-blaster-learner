package service_test

import "github.com/drradao/zigbee2mqtt-ir-blaster-learner/internal/mqtt"

type publishCall struct {
	topic   string
	payload []byte
}

type mockMQTTClient struct {
	publishCalls []publishCall
	publishErr   error
	connected    bool
}

func (m *mockMQTTClient) Connect() error                                  { return nil }
func (m *mockMQTTClient) Disconnect()                                     {}
func (m *mockMQTTClient) Subscribe(_ string, _ mqtt.MessageHandler) error { return nil }
func (m *mockMQTTClient) IsConnected() bool                               { return m.connected }
func (m *mockMQTTClient) SetStatusChannel(_ chan<- bool)                  {}

func (m *mockMQTTClient) Publish(topic string, payload []byte) error {
	m.publishCalls = append(m.publishCalls, publishCall{topic: topic, payload: payload})
	return m.publishErr
}
