// Package session defines the persistent session model and the repository
// interface for loading and saving sessions to YAML files.
package session

import "time"

// Command is a named IR command stored in a session.
type Command struct {
	Name       string    `yaml:"name"`
	Payload    string    `yaml:"payload"` // base64-encoded raw bytes
	CapturedAt time.Time `yaml:"captured_at"`
}

// Session groups commands for a single device.
type Session struct {
	Device   string    `yaml:"device"`
	Commands []Command `yaml:"commands"`
}
