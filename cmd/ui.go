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
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/config"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/flags"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/mqtt"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/service"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/session"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the interactive TUI (default interface)",
	RunE:  runUI,
}

func runUI(cmd *cobra.Command, _ []string) error {
	fv := buildFlagValues(cmd)
	logger := newLogger(fv.LogFile)

	// mqttCh bridges the MQTT subscriber goroutine and the bubbletea event
	// loop. A buffer of 32 avoids blocking the subscriber on bursts.
	mqttCh := make(chan buffer.IRMessage, 32)
	buf := buffer.New()

	var (
		mqttClient  mqtt.MQTTClient
		cfg         *config.Config
		sessionRepo session.SessionRepository
	)

	app := fx.New(
		fx.Supply(fv),
		fx.Supply(logger),
		config.Module,
		mqtt.Module,
		session.Module,
		fx.Populate(&mqttClient, &cfg, &sessionRepo),
		fx.NopLogger,
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startCancel()
	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("startup failed: %w", err)
	}

	listenTopic := mqtt.ListenTopic(cfg.Device)
	publishTopic := mqtt.PublishTopic(cfg.Device)

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

// Ensure flags import is used.
var _ flags.FlagValues
