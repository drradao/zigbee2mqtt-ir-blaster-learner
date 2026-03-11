package mqtt

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/config"
)

// Module is the FX module that provides MQTTClient and manages its lifecycle.
var Module = fx.Module("mqtt",
	fx.Provide(newMQTTClient),
	fx.Invoke(registerLifecycle),
)

func newMQTTClient(cfg *config.Config, logger *slog.Logger) MQTTClient {
	return newPahoClient(cfg.MQTT.Host, cfg.MQTT.Port, cfg.MQTT.User, cfg.MQTT.Password, logger)
}

func registerLifecycle(lc fx.Lifecycle, client MQTTClient, logger *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("connecting to MQTT broker")
			return client.Connect()
		},
		OnStop: func(_ context.Context) error {
			logger.Info("disconnecting from MQTT broker")
			client.Disconnect()
			return nil
		},
	})
}
