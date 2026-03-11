package session

import (
	"go.uber.org/fx"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/config"
)

// Module is the FX module that provides SessionRepository.
var Module = fx.Module("session",
	fx.Provide(newSessionRepository),
)

func newSessionRepository(cfg *config.Config) SessionRepository {
	path := cfg.Session
	if path == "" {
		path = "session.yaml"
	}
	return newYAMLRepository(path)
}
