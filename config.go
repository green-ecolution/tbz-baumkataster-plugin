package main

import (
	"github.com/caarlos0/env/v11"
	"net/url"
	"time"
)

type Config struct {
	SyncInterval time.Duration `env:"GE_SYNC_INTERVAL"`
	PluginSlug   string        `env:"GE_PLUGIN_SLUG"`
	PluginPath   *url.URL      `env:"GE_PLUGIN_PATH"`
	PluginPort   int           `env:"GE_PLUGIN_PORT"`
	HostPath     *url.URL      `env:"GE_HOST_PATH"`
	ClientID     string        `env:"GE_CLIENT_ID"`
	ClientSecret string        `env:"GE_CLIENT_SECRET"`
	DbURL        string        `env:"GE_DB_URL"`
}

func ParseConfig() Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	return cfg
}
