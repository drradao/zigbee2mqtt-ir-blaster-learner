// Package cmd implements the app CLI using cobra. All subcommands live in
// this package. Persistent flags are parsed here and forwarded to the FX
// container as a flags.FlagValues value object.
package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/drradao/zigbee2mqtt-ir-blaster-learner/internal/flags"
)

var rootCmd = &cobra.Command{
	Use:   "zigbee2mqtt-ir-blaster-learner",
	Short: "Learn and replay IR codes via zigbee2mqtt",
}

// globalFlags holds the raw persistent flag values before they are packaged
// into a flags.FlagValues and supplied to FX.
var globalFlags flags.FlagValues

// Execute is the main entry point called from main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globalFlags.ConfigPath, "config", "", "config file path (default: ~/.config/zigbee2mqtt-ir-blaster-learner/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.MQTTHost, "mqtt-host", "", "MQTT broker host")
	rootCmd.PersistentFlags().IntVar(&globalFlags.MQTTPort, "mqtt-port", 0, "MQTT broker port")
	rootCmd.PersistentFlags().StringVar(&globalFlags.MQTTUser, "mqtt-user", "", "MQTT username")
	rootCmd.PersistentFlags().StringVar(&globalFlags.MQTTPassword, "mqtt-password", "", "MQTT password")
	rootCmd.PersistentFlags().StringVar(&globalFlags.Device, "device", "", "zigbee2mqtt device friendly name")
	rootCmd.PersistentFlags().StringVar(&globalFlags.LogFile, "log", "", "log file path (default: stderr)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.Session, "session", "", "session YAML file path (default: session.yaml)")
	rootCmd.AddCommand(captureCmd)
	rootCmd.AddCommand(uiCmd)
}

// buildFlagValues copies globalFlags and records which flags were explicitly
// set so that config/loader.go can apply only intentional overrides.
func buildFlagValues(cmd *cobra.Command) flags.FlagValues {
	fv := globalFlags
	fv.Changed = make(map[string]bool)
	cmd.Flags().Visit(func(f *pflag.Flag) {
		fv.Changed[f.Name] = true
	})
	cmd.InheritedFlags().Visit(func(f *pflag.Flag) {
		fv.Changed[f.Name] = true
	})
	return fv
}

// newLogger builds an slog.Logger that writes to stderr (plain text, debug
// level) or to a file (JSON, debug level) when --log is set.
// The returned io.Closer must be closed by the caller to flush buffered writes.
func newLogger(logFile string) (*slog.Logger, io.Closer) {
	if logFile == "" {
		return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})), io.NopCloser(nil)
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open log file %s: %v\n", logFile, err)
		return slog.New(slog.NewTextHandler(os.Stderr, nil)), io.NopCloser(nil)
	}
	return slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})), f
}
