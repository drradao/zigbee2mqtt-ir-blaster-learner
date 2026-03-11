// Package buffer holds the in-memory ring of raw IR messages received from
// the MQTT broker during a session. It is safe for concurrent use.
package buffer

import (
	"sync"
	"time"
)

// IRMessage is a single decoded MQTT payload that contained a learned_ir_code.
type IRMessage struct {
	// Payload is the raw base64 string extracted from the learned_ir_code field.
	Payload    []byte
	Topic      string
	ReceivedAt time.Time
}

// IRBuffer is a thread-safe ordered list of IR messages.
type IRBuffer struct {
	mu   sync.RWMutex
	msgs []IRMessage
}

// New returns an empty IRBuffer.
func New() *IRBuffer {
	return &IRBuffer{}
}

// Add appends msg to the buffer.
func (b *IRBuffer) Add(msg IRMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.msgs = append(b.msgs, msg)
}

// Remove deletes the message at index idx. Out-of-range indices are ignored.
func (b *IRBuffer) Remove(idx int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if idx < 0 || idx >= len(b.msgs) {
		return
	}
	b.msgs = append(b.msgs[:idx], b.msgs[idx+1:]...)
}

// Clear removes all messages.
func (b *IRBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.msgs = nil
}

// All returns a snapshot of all messages.
func (b *IRBuffer) All() []IRMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]IRMessage, len(b.msgs))
	copy(result, b.msgs)
	return result
}

// Len returns the number of buffered messages.
func (b *IRBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.msgs)
}
