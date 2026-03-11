package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/buffer"
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
	logger := newLogger(fv.LogFile)

	cfg, cfgPath, err := runPreFlight(fv)
	if err != nil {
		return err
	}

	if uiLogin {
		if err := runLoginPreflight(cfg, cfgPath); err != nil {
			return err
		}
	}

	// mqttCh bridges the MQTT subscriber goroutine and the bubbletea event
	// loop. A buffer of 32 avoids blocking the subscriber on bursts.
	mqttCh := make(chan buffer.IRMessage, 32)
	// statusCh carries connection state changes (true=connected, false=lost)
	// from the paho callbacks into the bubbletea event loop.
	statusCh := make(chan bool, 4)
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
	if err := mqttClient.Subscribe(listenTopic, buildMQTTHandler(buf, mqttCh, logger)); err != nil {
		stopFX(app)
		return fmt.Errorf("subscribe failed: %w", err)
	}

	learnSvc := service.NewLearnService(mqttClient, publishTopic)
	replaySvc := service.NewReplayService(mqttClient, publishTopic)

	sessionFile := cfg.Session
	if sessionFile == "" {
		sessionFile = "session.yaml"
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
		stopFX(app)
		return fmt.Errorf("TUI error: %w", err)
	}

	stopFX(app)
	return nil
}

// buildMQTTHandler returns an mqtt.MessageHandler that filters for
// learned_ir_code payloads and forwards them to both the in-memory buffer
// and the bubbletea channel.
func buildMQTTHandler(buf *buffer.IRBuffer, ch chan buffer.IRMessage, logger *slog.Logger) mqtt.MessageHandler {
	return func(topic string, payload []byte) {
		code := extractCode(payload, logger)
		if code == "" {
			return
		}
		msg := buffer.IRMessage{
			Payload:    []byte(code),
			Topic:      topic,
			ReceivedAt: time.Now(),
		}
		buf.Add(msg)
		select {
		case ch <- msg:
		default:
			logger.Warn("mqtt channel full, dropping IR message")
		}
	}
}
