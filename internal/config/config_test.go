package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/flags"
)

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MQTT.Port != 1883 {
		t.Errorf("expected default port 1883, got %d", cfg.MQTT.Port)
	}
	if cfg.MQTT.Host != "" || cfg.MQTT.User != "" || cfg.MQTT.Password != "" || cfg.Device != "" {
		t.Error("expected all string fields empty for default config")
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	content := `
mqtt:
  host: broker.example.com
  port: 8883
  user: alice
  password: secret
device: my-ir-blaster
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MQTT.Host != "broker.example.com" {
		t.Errorf("host: got %q", cfg.MQTT.Host)
	}
	if cfg.MQTT.Port != 8883 {
		t.Errorf("port: got %d", cfg.MQTT.Port)
	}
	if cfg.MQTT.User != "alice" {
		t.Errorf("user: got %q", cfg.MQTT.User)
	}
	if cfg.MQTT.Password != "secret" {
		t.Errorf("password: got %q", cfg.MQTT.Password)
	}
	if cfg.Device != "my-ir-blaster" {
		t.Errorf("device: got %q", cfg.Device)
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(":\t:bad yaml{{"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestLoad_PortZeroAppliesDefault(t *testing.T) {
	content := "mqtt:\n  port: 0\n"
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MQTT.Port != 1883 {
		t.Errorf("expected default port 1883 for port=0, got %d", cfg.MQTT.Port)
	}
}

func TestLoad_ExplicitPort(t *testing.T) {
	content := "mqtt:\n  port: 9999\n"
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MQTT.Port != 9999 {
		t.Errorf("expected port 9999, got %d", cfg.MQTT.Port)
	}
}

func TestSave_CreatesFileAndDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "config.yaml")
	cfg := &Config{MQTT: MQTTConfig{Host: "h", Port: 1883}}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestSave_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	original := &Config{
		MQTT:   MQTTConfig{Host: "broker", Port: 1883, User: "u", Password: "p"},
		Device: "dev1",
	}
	if err := Save(path, original); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.MQTT.Host != original.MQTT.Host {
		t.Errorf("host: got %q, want %q", loaded.MQTT.Host, original.MQTT.Host)
	}
	if loaded.MQTT.Port != original.MQTT.Port {
		t.Errorf("port: got %d, want %d", loaded.MQTT.Port, original.MQTT.Port)
	}
	if loaded.MQTT.User != original.MQTT.User {
		t.Errorf("user: got %q, want %q", loaded.MQTT.User, original.MQTT.User)
	}
	if loaded.MQTT.Password != original.MQTT.Password {
		t.Errorf("password: got %q, want %q", loaded.MQTT.Password, original.MQTT.Password)
	}
	if loaded.Device != original.Device {
		t.Errorf("device: got %q, want %q", loaded.Device, original.Device)
	}
}

func TestSave_TransientFieldsNotPersisted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := &Config{
		MQTT:    MQTTConfig{Host: "h", Port: 1883},
		LogFile: "/var/log/irlearn.log",
		Session: "/tmp/session.yaml",
	}
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "log") {
		t.Errorf("LogFile should not be in saved YAML, but found 'log' in: %s", content)
	}
	if strings.Contains(content, "session") {
		t.Errorf("Session should not be in saved YAML, but found 'session' in: %s", content)
	}
}

func TestApplyFlagOverrides_AllChanged(t *testing.T) {
	cfg := &Config{MQTT: MQTTConfig{Host: "old", Port: 1883, User: "old", Password: "old"}, Device: "old"}
	fv := flags.FlagValues{
		MQTTHost:     "new-host",
		MQTTPort:     9999,
		MQTTUser:     "new-user",
		MQTTPassword: "new-pass",
		Device:       "new-dev",
		Changed: map[string]bool{
			"mqtt-host":     true,
			"mqtt-port":     true,
			"mqtt-user":     true,
			"mqtt-password": true,
			"device":        true,
		},
	}
	ApplyFlagOverrides(cfg, fv)
	if cfg.MQTT.Host != "new-host" {
		t.Errorf("host: got %q", cfg.MQTT.Host)
	}
	if cfg.MQTT.Port != 9999 {
		t.Errorf("port: got %d", cfg.MQTT.Port)
	}
	if cfg.MQTT.User != "new-user" {
		t.Errorf("user: got %q", cfg.MQTT.User)
	}
	if cfg.MQTT.Password != "new-pass" {
		t.Errorf("password: got %q", cfg.MQTT.Password)
	}
	if cfg.Device != "new-dev" {
		t.Errorf("device: got %q", cfg.Device)
	}
}

func TestApplyFlagOverrides_NoneChanged(t *testing.T) {
	cfg := &Config{MQTT: MQTTConfig{Host: "orig", Port: 1883, User: "orig", Password: "orig"}, Device: "orig"}
	fv := flags.FlagValues{
		MQTTHost: "ignored",
		Changed:  map[string]bool{},
	}
	ApplyFlagOverrides(cfg, fv)
	if cfg.MQTT.Host != "orig" {
		t.Errorf("host should not change, got %q", cfg.MQTT.Host)
	}
	if cfg.Device != "orig" {
		t.Errorf("device should not change, got %q", cfg.Device)
	}
}

func TestApplyFlagOverrides_PartialChanged(t *testing.T) {
	cfg := &Config{MQTT: MQTTConfig{Host: "orig", Port: 1883, User: "orig"}, Device: "orig"}
	fv := flags.FlagValues{
		MQTTHost: "new-host",
		MQTTUser: "ignored",
		Changed:  map[string]bool{"mqtt-host": true},
	}
	ApplyFlagOverrides(cfg, fv)
	if cfg.MQTT.Host != "new-host" {
		t.Errorf("host: got %q, want new-host", cfg.MQTT.Host)
	}
	if cfg.MQTT.User != "orig" {
		t.Errorf("user should remain 'orig', got %q", cfg.MQTT.User)
	}
	if cfg.Device != "orig" {
		t.Errorf("device should remain 'orig', got %q", cfg.Device)
	}
}
