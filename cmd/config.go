package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/config"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui/setup"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit irlearn configuration interactively",
	RunE:  runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, _ []string) error {
	fv := buildFlagValues(cmd)
	path := fv.ConfigPath
	if path == "" {
		path = config.DefaultConfigPath()
	}

	existing, _ := config.Load(path)
	if existing == nil {
		existing = &config.Config{}
	}
	config.ApplyFlagOverrides(existing, fv)

	result, err := setup.Run(*existing, setup.ModeFull, path)
	if err != nil {
		return fmt.Errorf("config form: %w", err)
	}
	if result.Aborted {
		fmt.Println("Config edit cancelled.")
		return nil
	}
	if result.Err != nil {
		return fmt.Errorf("saving config: %w", result.Err)
	}
	if result.Saved {
		fmt.Printf("Config saved to %s\n", path)
	} else {
		fmt.Println("Config not saved.")
	}
	return nil
}
