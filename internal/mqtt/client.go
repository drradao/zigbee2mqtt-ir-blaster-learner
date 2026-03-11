// Package mqtt wraps the paho MQTT client behind an interface so that the rest
// of the application can depend on the interface rather than the concrete
// implementation, making services testable without a real broker.
package mqtt

// MessageHandler is called whenever a message arrives on a subscribed topic.
type MessageHandler func(topic string, payload []byte)

// MQTTClient is the interface the application uses for all broker interactions.
type MQTTClient interface {
	// Connect establishes the connection to the broker. It is called once
	// during the FX OnStart lifecycle hook.
	Connect() error
	// Disconnect closes the connection gracefully. Called during OnStop.
	Disconnect()
	// Subscribe registers handler for the given topic filter.
	Subscribe(topic string, handler MessageHandler) error
	// Publish sends payload to topic with QoS 1.
	Publish(topic string, payload []byte) error
	// IsConnected reports whether the underlying connection is live.
	IsConnected() bool
	// SetStatusChannel registers a channel that receives true on connect and
	// false on connection loss. The send is non-blocking; use a buffered channel.
	SetStatusChannel(ch chan<- bool)
}
