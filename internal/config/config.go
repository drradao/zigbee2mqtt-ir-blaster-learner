// Package config handles loading and merging configuration from a YAML file
// and CLI flag overrides.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// MQTTConfig holds MQTT broker connection settings.
type MQTTConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// Config is the top-level application configuration.
type Config struct {
	MQTT    MQTTConfig `yaml:"mqtt"`
	Device  string     `yaml:"device"`
	LogFile string     `yaml:"-"`
	Session string     `yaml:"-"`
}

// DefaultConfigPath returns the default config file location:
// ~/.config/mqttirlearn/config.yaml
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return home + "/.config/mqttirlearn/config.yaml"
}

// Load reads a YAML config from path. A missing file is not an error — it
// returns a zero-value Config so the caller can apply flag overrides on top.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
