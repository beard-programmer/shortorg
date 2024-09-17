package app

import (
	"fmt"

	apiServer "github.com/beard-programmer/shortorg/internal/api"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/cache"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/postgres"
	"github.com/spf13/viper"
)

type config struct {
	Env                string
	EncodedUrlsQueSize int
	Concurrency        int
	IsDebug            bool
	PostgresClients    postgres.ClientsConfig `mapstructure:"PostgresClients"`
	Cache              cache.Config           `mapstructure:"Cache"`
	Infrastructure     infrastructure.Config  `mapstructure:"Infrastructure"`
	APIServer          apiServer.Config       `mapstructure:"APIServer"`
}

func (config) load(env string) (*config, error) {
	viperConfig := viper.New()

	viperConfig.SetConfigType("toml")
	viperConfig.AddConfigPath("./config/")
	viperConfig.SetConfigName(fmt.Sprintf("application.%s", env))

	if err := viperConfig.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading cfg file: %w", err)
	}

	var cfg config

	if err := viperConfig.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &cfg, nil
}

func (c config) isProdEnv() bool {
	return c.Env == "prod"
}

func (c config) isTestEnv() bool {
	return c.Env == "test"
}

func (c config) isDevEnv() bool {
	return c.Env == "dev"
}
