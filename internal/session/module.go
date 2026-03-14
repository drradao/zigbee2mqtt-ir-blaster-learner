package session

import (
	"go.uber.org/fx"

	"github.com/drradao/zigbee2mqtt-ir-blaster-learner/internal/config"
)

// Module is the FX module that provides SessionRepository.
var Module = fx.Module("session",
	fx.Provide(newSessionRepository),
)

func newSessionRepository(cfg *config.Config) (SessionRepository, error) {
	path := cfg.Session
	if path == "" {
		var err error
		path, err = config.DefaultSessionPath()
		if err != nil {
			return nil, err
		}
	}
	return newYAMLRepository(path), nil
}
