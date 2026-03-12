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
}

// LoginSubmitted is dispatched when the login modal collects credentials.
type LoginSubmitted struct{ User, Password string }

// LoginDismissed is dispatched when the login modal is closed without submitting.
type LoginDismissed struct{}
