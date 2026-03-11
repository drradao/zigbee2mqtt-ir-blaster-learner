package config

import (
	"go.uber.org/fx"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/flags"
)

// Module is the FX module that provides *Config to the container.
var Module = fx.Module("config",
	fx.Provide(newConfig),
)

func newConfig(fv flags.FlagValues) (*Config, error) {
	path := fv.ConfigPath
	if path == "" {
		path = DefaultConfigPath()
	}
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}
	ApplyFlagOverrides(cfg, fv)
	return cfg, nil
}
