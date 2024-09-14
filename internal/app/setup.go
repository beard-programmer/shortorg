package app

import (
	"context"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/beard-programmer/shortorg/internal/base58"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/cache"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/postgresClients"
	"github.com/beard-programmer/shortorg/internal/decode"
	decodeInfrastructure "github.com/beard-programmer/shortorg/internal/decode/infrastructure"
	"github.com/beard-programmer/shortorg/internal/encode"
	encodeInfrastructure "github.com/beard-programmer/shortorg/internal/encode/infrastructure"
	"github.com/kelseyhightower/envconfig"
)

func (app *App) Setup(ctx context.Context) error {
	err := app.setupConfig(ctx)
	if err != nil {
		return fmt.Errorf("setup config: %w", err)
	}
	//
	//err = app.setupContextUtils(ctx)
	//if err != nil {
	//	return fmt.Errorf("setup context utils: %w", err)
	//}

	err = app.setupPostgresClients(ctx)
	if err != nil {
		return fmt.Errorf("setup postgres clients: %w", err)
	}

	err = app.setupCache(ctx)
	if err != nil {
		return fmt.Errorf("setup cache: %w", err)
	}

	err = app.setupEncodedUrlStore(ctx)
	if err != nil {
		return fmt.Errorf("setup encoded url store: %w", err)
	}

	err = app.setupTokenKeyStore(ctx)
	if err != nil {
		return fmt.Errorf("setup token key store: %w", err)
	}

	err = app.setupUseCaseFns(ctx)
	if err != nil {
		return fmt.Errorf("setup use cases: %w", err)
	}

	err = app.setupEventHandlers(ctx)
	if err != nil {
		return fmt.Errorf("setup event handlers: %w", err)
	}

	return nil
}

func (app *App) setupConfig(ctx context.Context) error {
	var config Config

	_, err := toml.DecodeFile("./config/config.dev.toml", &config)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var envConfig Config
	err = envconfig.Process("", &envConfig)
	if err != nil {
		return fmt.Errorf("error processing environment variables: %w", err)
	}

	if envConfig.IsProdEnv() {
		config = envConfig
	}

	//// 3. Validate the configuration
	//err = validateConfig(&config)
	//if err != nil {
	//	return nil, fmt.Errorf("configuration validation error: %w", err)
	//}

	app.config = config

	return nil
}

func (app *App) setupPostgresClients(ctx context.Context) error {
	clients, err := postgresClients.New(ctx, app.logger, app.config.PostgresClientsConfig, app.Name(), app.config.IsProdEnv())
	if err != nil {
		return fmt.Errorf("setupPostgresClients: %w", err)
	}

	app.postgresClients = clients
	return nil
}

func (app *App) setupCache(ctx context.Context) error {
	inMemory, err := cache.NewInMemory[string](app.config.CacheConfig)
	if err != nil {
		return fmt.Errorf("setupCache: %w", err)
	}

	app.cache = inMemory
	return nil
}

func (app *App) setupEncodedUrlStore(ctx context.Context) error {
	store, err := infrastructure.NewEncodedUrlStore(app.postgresClients.ShortorgClient, app.cache)
	if err != nil {
		return fmt.Errorf("setupEncodedUrlStore: %w", err)
	}
	app.encodedUrlStore = store
	return nil
}

func (app *App) setupTokenKeyStore(ctx context.Context) error {
	store, err := infrastructure.NewTokenKeyStore(ctx, app.postgresClients.TokenIdentifierClient, app.logger, 10000)
	if err != nil {
		return fmt.Errorf("setupTokenKeyStore: %w", err)
	}
	app.tokenKeyStore = store
	return nil
}

func (app *App) setupUseCaseFns(ctx context.Context) error {
	app.urlWasEncodedChan = make(chan encode.UrlWasEncoded, app.config.EncodedUrlsQueSize)
	app.encodeFn = encode.NewEncodeFn(app.tokenKeyStore, infrastructure.UrlParser{}, base58.Codec{}, app.logger, app.urlWasEncodedChan)
	app.decodeFn = decode.NewDecodeFn(app.logger, decodeInfrastructure.UrlParser{}, base58.Codec{}, app.encodedUrlStore)
	return nil
}

func (app *App) setupEventHandlers(ctx context.Context) error {
	app.urlWasEncodedHandler = encodeInfrastructure.NewUrlWasEncodedHandler(app.logger, app.encodedUrlStore, 10000, 1, 250*time.Millisecond, app.urlWasEncodedChan)
	return nil
}
