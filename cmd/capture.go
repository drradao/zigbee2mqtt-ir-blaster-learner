package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/config"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/flags"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/mqtt"
)

var captureTimeout int
var captureVerbose bool

var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Headless single-shot IR code capture — prints the learned_ir_code to stdout",
	RunE:  runCapture,
}

func init() {
	captureCmd.Flags().IntVar(&captureTimeout, "timeout", 30, "seconds to wait for an IR code")
	captureCmd.Flags().BoolVar(&captureVerbose, "verbose", false, "print progress messages to stderr")
}

func runCapture(cmd *cobra.Command, _ []string) error {
	fv := buildFlagValues(cmd)
	logger := newLogger(fv.LogFile)

	resultCh := make(chan string, 1)

	var (
		mqttClient mqtt.MQTTClient
		cfg        *config.Config
	)

	app := fx.New(
		fx.Supply(fv),
		fx.Supply(logger),
		config.Module,
		mqtt.Module,
		fx.Populate(&mqttClient, &cfg),
		fx.NopLogger,
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startCancel()
	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("startup failed: %w", err)
	}

	listenTopic := mqtt.ListenTopic(cfg.Device)
	publishTopic := mqtt.PublishTopic(cfg.Device)

	if err := mqttClient.Subscribe(listenTopic, func(_ string, payload []byte) {
		code := extractCode(payload, logger)
		if code != "" {
			select {
			case resultCh <- code:
			default:
			}
		}
	}); err != nil {
		stopFX(app)
		return fmt.Errorf("subscribe failed: %w", err)
	}

	learnPayload, _ := json.Marshal(map[string]interface{}{"learn_ir_code": true})
	if err := mqttClient.Publish(publishTopic, learnPayload); err != nil {
		stopFX(app)
		return fmt.Errorf("publish learn command failed: %w", err)
	}

	if captureVerbose {
		fmt.Fprintln(os.Stderr, "Waiting for IR code...")
	}

	timeout := time.Duration(captureTimeout) * time.Second
	var exitCode int
	select {
	case code := <-resultCh:
		fmt.Println(code)
	case <-time.After(timeout):
		if captureVerbose {
			fmt.Fprintln(os.Stderr, "Timeout: no IR code received")
		}
		exitCode = 1
	}

	stopFX(app)
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}

func extractCode(payload []byte, logger *slog.Logger) string {
	var m map[string]interface{}
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Debug("failed to unmarshal MQTT payload", "error", err)
		return ""
	}
	v, ok := m["learned_ir_code"]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func stopFX(app *fx.App) {
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = app.Stop(stopCtx)
}

// Ensure flags import is used through the FlagValues type used in buildFlagValues.
var _ flags.FlagValues
