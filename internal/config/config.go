// Package config handles loading and merging configuration from a YAML file
// and CLI flag overrides.
package config

import (
	"fmt"
	"os"
	"path/filepath"

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
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "mqttirlearn", "config.yaml"), nil
}

// DefaultSessionPath returns the default session file location:
// ~/.config/mqttirlearn/session.yaml
func DefaultSessionPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "mqttirlearn", "session.yaml"), nil
}

// Load reads a YAML config from path. A missing file is not an error — it
// returns a Config with defaults applied so the caller can apply flag overrides
// on top. Port defaults to 1883 when not explicitly set.
func Load(path string) (*Config, error) {
	cfg := &Config{MQTT: MQTTConfig{Port: 1883}}
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
	if cfg.MQTT.Port == 0 {
		cfg.MQTT.Port = 1883
	}
	return cfg, nil
}

// Save writes cfg atomically to path via a temp file rename. Fields tagged
// yaml:"-" (LogFile, Session) are not written.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "irlearn-cfg-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
