package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/buffer"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/config"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/mqtt"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/service"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/session"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui"
)

var uiLogin bool

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the interactive TUI (default interface)",
	RunE:  runUI,
}

func init() {
	uiCmd.Flags().BoolVar(&uiLogin, "login", false, "collect MQTT credentials before connecting (not saved to config)")
}

func runUI(cmd *cobra.Command, _ []string) error {
	fv := buildFlagValues(cmd)
	logger, logCloser := newLogger(fv.LogFile)
	defer logCloser.Close()

	cfg, cfgPath, err := runPreFlight(fv)
	if err != nil {
		if errors.Is(err, ErrSetupAborted) {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			os.Exit(1)
		}
		return err
	}

	if uiLogin {
		if err := runLoginPreflight(cfg, cfgPath); err != nil {
			if errors.Is(err, ErrSetupAborted) {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				os.Exit(1)
			}
			return err
		}
	}

	// mqttCh bridges the MQTT subscriber goroutine and the bubbletea event
	// loop. A buffer of 32 avoids blocking the subscriber on bursts.
	mqttCh := make(chan buffer.IRMessage, 32)
	// statusCh carries connection state changes (true=connected, false=lost)
	// from the paho callbacks into the bubbletea event loop.
	statusCh := make(chan bool, 4)
	// handlerStopped prevents the MQTT handler from sending on closed channels
	// after teardown has begun.
	var handlerStopped atomic.Bool
	defer close(mqttCh)
	defer close(statusCh)
	buf := buffer.New()

	var (
		mqttClient  mqtt.MQTTClient
		sessionRepo session.SessionRepository
	)

	// Supply the pre-flight resolved config directly instead of config.Module,
	// so that any credentials or values collected by the setup form are applied.
	app := fx.New(
		fx.Supply(fv),
		fx.Supply(logger),
		fx.Supply(cfg),
		mqtt.Module,
		session.Module,
		fx.Populate(&mqttClient, &sessionRepo),
		fx.NopLogger,
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startCancel()
	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("startup failed: %w", err)
	}

	listenTopic := mqtt.ListenTopic(cfg.Device)
	publishTopic := mqtt.PublishTopic(cfg.Device)

	// Bridge paho connection state changes into the TUI status channel.
	mqttClient.SetStatusChannel(statusCh)

	// Subscribe forwards filtered MQTT messages into the channel consumed by the TUI.
	if err := mqttClient.Subscribe(listenTopic, buildMQTTHandler(&handlerStopped, mqttCh, logger)); err != nil {
		handlerStopped.Store(true)
		stopFX(app)
		return fmt.Errorf("subscribe failed: %w", err)
	}

	learnSvc := service.NewLearnService(mqttClient, publishTopic)
	replaySvc := service.NewReplayService(mqttClient, publishTopic)

	sessionFile := cfg.Session
	if sessionFile == "" {
		var err error
		sessionFile, err = config.DefaultSessionPath()
		if err != nil {
			stopFX(app)
			return fmt.Errorf("resolving session path: %w", err)
		}
	}

	model := tui.NewModel(tui.ModelParams{
		Buf:         buf,
		MQTTClient:  mqttClient,
		LearnSvc:    learnSvc,
		ReplaySvc:   replaySvc,
		SessionRepo: sessionRepo,
		Logger:      logger,
		SessionFile: sessionFile,
		MQTTCh:      mqttCh,
		StatusCh:    statusCh,
	})

	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		handlerStopped.Store(true)
		stopFX(app)
		return fmt.Errorf("TUI error: %w", err)
	}

	handlerStopped.Store(true)
	stopFX(app)
	return nil
}

// buildMQTTHandler returns an mqtt.MessageHandler that filters for
// learned_ir_code payloads and forwards them to the bubbletea channel.
// The TUI Update loop is the sole owner responsible for adding messages to the buffer.
// stopped must be set to true before the channel is closed to prevent send-on-closed-channel panics.
func buildMQTTHandler(stopped *atomic.Bool, ch chan buffer.IRMessage, logger *slog.Logger) mqtt.MessageHandler {
	return func(topic string, payload []byte) {
		code := extractCode(payload, logger)
		if code == "" {
			return
		}
		if stopped.Load() {
			return
		}
		msg := buffer.IRMessage{
			Payload:    []byte(code),
			Topic:      topic,
			ReceivedAt: time.Now(),
		}
		select {
		case ch <- msg:
		default:
			logger.Warn("mqtt channel full, dropping IR message")
		}
	}
}
