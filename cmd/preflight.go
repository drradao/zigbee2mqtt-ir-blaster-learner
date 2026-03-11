package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/config"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/flags"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui/setup"
)

// ErrSetupAborted is returned when the user cancels the setup or login form.
var ErrSetupAborted = errors.New("setup cancelled by user")

// runPreFlight loads config, applies flag overrides, and runs the setup form
// if required fields (Host or Device) are empty.
// Returns the resolved *config.Config and the active config path.
// Returns ErrSetupAborted if the user aborts the form.
func runPreFlight(fv flags.FlagValues) (*config.Config, string, error) {
	path := fv.ConfigPath
	if path == "" {
		var err error
		path, err = config.DefaultConfigPath()
		if err != nil {
			return nil, "", fmt.Errorf("resolving config path: %w", err)
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		return nil, "", fmt.Errorf("loading config: %w", err)
	}
	config.ApplyFlagOverrides(cfg, fv)

	if cfg.MQTT.Host == "" || cfg.Device == "" {
		result, err := setup.Run(*cfg, setup.ModeMissing, path)
		if err != nil {
			return nil, "", fmt.Errorf("setup form: %w", err)
		}
		if result.Aborted {
			return nil, "", ErrSetupAborted
		}
		// Merge the form result back, preserving yaml:"-" fields.
		merged := result.Config
		merged.LogFile = cfg.LogFile
		merged.Session = cfg.Session
		cfg = &merged
	}

	return cfg, path, nil
}

// runLoginPreflight shows the credentials-only form before the main TUI starts.
// Used by runUI when --login is set.
func runLoginPreflight(cfg *config.Config, _ string) error {
	result, err := setup.Run(*cfg, setup.ModeLogin, "")
	if err != nil {
		return fmt.Errorf("login form: %w", err)
	}
	if result.Aborted {
		return ErrSetupAborted
	}
	cfg.MQTT.User = result.Config.MQTT.User
	cfg.MQTT.Password = result.Config.MQTT.Password
	return nil
}

// headlessLoginPrompt prompts for credentials via stderr/stdin.
// Used by capture --login.
// Note: password input is visible (no echo suppression) to avoid introducing
// golang.org/x/term as a dependency.
func headlessLoginPrompt(cfg *config.Config) {
	if cfg.MQTT.User == "" {
		fmt.Fprint(os.Stderr, "MQTT username: ")
		fmt.Fscan(os.Stdin, &cfg.MQTT.User)
	}
	if cfg.MQTT.Password == "" {
		fmt.Fprint(os.Stderr, "MQTT password: ")
		fmt.Fscan(os.Stdin, &cfg.MQTT.Password)
	}
}
