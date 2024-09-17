package app

import (
	apiServer "github.com/beard-programmer/shortorg/internal/api"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/cache"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/postgres"
)

type config struct {
	Env                 string
	TokenKeyStoreBuffer int
	EncodedUrlsQueSize  int
	Concurrency         int
	IsDebug             bool
	UseCache            bool
	APIServer           apiServer.Config       `mapstructure:"APIServer"`
	PostgresClients     postgres.ClientsConfig `mapstructure:"PostgresClients"`
	Cache               cache.Config           `mapstructure:"Cache"`
}

func (c config) isProdEnv() bool {
	return c.Env == "production"
}
