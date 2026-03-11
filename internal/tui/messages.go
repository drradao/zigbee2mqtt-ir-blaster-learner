// Package tui implements the bubbletea-based terminal UI.
package tui

import "github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/buffer"

// MQTTMessageReceived is dispatched to the bubbletea event loop when a new
// IR code arrives from the broker.
type MQTTMessageReceived struct {
	Msg buffer.IRMessage
}

// StatusChanged updates the MQTT connection status shown in the status bar.
type StatusChanged struct {
	Connected bool
	Text      string
}
