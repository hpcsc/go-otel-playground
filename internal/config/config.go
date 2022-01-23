package config

import "github.com/caarlos0/env/v6"

type Configuration struct {
	CollectorAddr string `env:"COLLECTOR_ADDR"`
	Port          int    `env:"PORT"`
}

func Load() (*Configuration, error) {
	var c Configuration
	if err := env.Parse(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
