package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ListenAddr         int    `envconfig:"APP_PORT" default:"8080"`
	OSRMBaseURL        string `envconfig:"OSRM_BASE_URL" required:"true"`
	OSRMTimeoutMs      int    `envconfig:"OSRM_TIMEOUT_MS" required:"true"`
	OSRMMaxConcurrency int    `envconfig:"OSRM_MAX_CONCURRNCY" required:"true"`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}

	if cfg.OSRMTimeoutMs <= 0 {
		return Config{}, fmt.Errorf("OSRM_TIMEOUT_MS: must be positive, got %d", cfg.OSRMTimeoutMs)
	}
	if cfg.OSRMMaxConcurrency <= 0 {
		return Config{}, fmt.Errorf("OSRM_MAX_CONCURRNCY: must be positive, got %d", cfg.OSRMMaxConcurrency)
	}

	return cfg, nil
}
