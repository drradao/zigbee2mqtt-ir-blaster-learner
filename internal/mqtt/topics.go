package mqtt

import "fmt"

// ListenTopic returns the topic on which the device publishes its state,
// including learned_ir_code payloads.
func ListenTopic(device string) string {
	return fmt.Sprintf("zigbee2mqtt/%s", device)
}

// PublishTopic returns the set topic used to send commands to the device.
func PublishTopic(device string) string {
	return fmt.Sprintf("zigbee2mqtt/%s/set", device)
}
