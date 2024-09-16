package app

import (
	"context"
	"fmt"

	apiServer "github.com/beard-programmer/shortorg/internal/api"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/cache"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/postgresClients"
	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
	encodeInfrastructure "github.com/beard-programmer/shortorg/internal/encode/infrastructure"

	"go.uber.org/zap"
)

type App struct {
	logger               *zap.Logger
	postgresClients      *postgresClients.Clients
	cache                infrastructure.Cache
	encodedUrlStore      *infrastructure.EncodedUrlStore
	tokenKeyStore        *infrastructure.TokenKeyStore
	urlWasEncodedChan    chan encode.UrlWasEncoded
	encodeFn             encode.Fn
	decodeFn             decode.Fn
	urlWasEncodedHandler encodeInfrastructure.UrlWasEncodedHandlerFn
	config               Config
}

func New(l *zap.Logger) *App {
	return &App{logger: l}
}

func (App) Name() string {
	return "shortorg"
}

type Config struct {
	UseCache              bool                          `toml:"use_cache" envconfig:"USE_CACHE"`
	Concurrency           int                           `toml:"concurrency" envconfig:"GOMAXPROCS" default:"2"`
	EncodedUrlsQueSize    int64                         `toml:"encoded_urls_queue_size" envconfig:"ENCODED_URLS_QUEUE_SIZE" default:"1000"`
	Debug                 bool                          `toml:"debug" envconfig:"DEBUG" default:"false"`
	ENV                   string                        `toml:"env" envconfig:"ENV" default:"development"`
	ApiServerConfig       apiServer.Config              `toml:"api_server" envconfig:"API_SERVER"`
	PostgresClientsConfig postgresClients.ClientsConfig `toml:"postgres_clients" envconfig:"POSTGRES_CLIENTS"`
	CacheConfig           cache.Config                  `toml:"cache" envconfig:"CACHE"`
}

func (c Config) IsProdEnv() bool {
	return c.ENV == "production"
}

func (app *App) Serve(ctx context.Context) error {
	api := apiServer.New(
		app.encodeFn,
		app.decodeFn,
		app.urlWasEncodedHandler,
		app.logger,
		app.config.ApiServerConfig,
		app.Name(),
	)

	err := api.Serve(ctx)
	if err != nil {
		return fmt.Errorf("api server: %w", err)
	}

	return nil
}
