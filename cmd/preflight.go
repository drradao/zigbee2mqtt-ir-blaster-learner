package cmd

import (
	"fmt"
	"os"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/config"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/flags"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui/setup"
)

// runPreFlight loads config, applies flag overrides, and runs the setup form
// if required fields (Host or Device) are empty.
// Returns the resolved *config.Config and the active config path.
// Calls os.Exit(1) if the user aborts the form.
func runPreFlight(fv flags.FlagValues) (*config.Config, string, error) {
	path := fv.ConfigPath
	if path == "" {
		path = config.DefaultConfigPath()
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
			fmt.Fprintln(os.Stderr, "Setup cancelled.")
			os.Exit(1)
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
		fmt.Fprintln(os.Stderr, "Login cancelled.")
		os.Exit(1)
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
