package config

import "github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/flags"

// ApplyFlagOverrides copies flag values into cfg only for flags the user
// explicitly set on the command line.
func ApplyFlagOverrides(cfg *Config, fv flags.FlagValues) {
	if fv.Changed["mqtt-host"] {
		cfg.MQTT.Host = fv.MQTTHost
	}
	if fv.Changed["mqtt-port"] {
		cfg.MQTT.Port = fv.MQTTPort
	}
	if fv.Changed["mqtt-user"] {
		cfg.MQTT.User = fv.MQTTUser
	}
	if fv.Changed["mqtt-password"] {
		cfg.MQTT.Password = fv.MQTTPassword
	}
	if fv.Changed["device"] {
		cfg.Device = fv.Device
	}
	if fv.Changed["log"] {
		cfg.LogFile = fv.LogFile
	}
	if fv.Changed["session"] {
		cfg.Session = fv.Session
	}
}
