package app

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/beard-programmer/shortorg/internal/base58"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/cache"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/postgres"
	"github.com/beard-programmer/shortorg/internal/decode"
	decodeInfrastructure "github.com/beard-programmer/shortorg/internal/decode/infrastructure"
	"github.com/beard-programmer/shortorg/internal/encode"
	encodeInfrastructure "github.com/beard-programmer/shortorg/internal/encode/infrastructure"
	"github.com/spf13/viper"
)

func (app *App) setup(ctx context.Context) error {
	err := app.setupConfig(ctx)
	if err != nil {
		return fmt.Errorf("setup config: %w", err)
	}

	_ = runtime.GOMAXPROCS(app.config.Concurrency)

	err = app.setupPostgresClients(ctx)
	if err != nil {
		return fmt.Errorf("setup postgres clients: %w", err)
	}

	err = app.setupCache(ctx)
	if err != nil {
		return fmt.Errorf("setup cache: %w", err)
	}

	err = app.setupEncodedURLStore(ctx)
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

func (app *App) setupConfig(_ context.Context) error {

	env := os.Getenv("APP_ENV")
	viperConfig := viper.New()

	viperConfig.SetConfigType("toml")
	viperConfig.AddConfigPath("./config/")
	viperConfig.SetConfigName(fmt.Sprintf("application.%s", env))

	if err := viperConfig.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var cfg config

	if err := viperConfig.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("unable to decode into struct: %w", err)
	}
	app.logger.Sugar().Infow("Configs set up", "config", cfg)

	app.config = cfg

	return nil
}

func (app *App) setupPostgresClients(ctx context.Context) error {
	clients, err := postgres.New(
		ctx,
		app.logger,
		app.config.PostgresClients,
		app.Name(),
		app.config.isProdEnv(),
	)
	if err != nil {
		return fmt.Errorf("setupPostgresClients: %w", err)
	}

	app.postgresClients = clients
	return nil
}

func (app *App) setupCache(_ context.Context) error {
	if !app.config.UseCache {
		app.cache = &cache.MockCache[string]{}

		return nil
	}

	inMemory, err := cache.NewInMemory[string](app.config.Cache)
	if err != nil {
		return fmt.Errorf("setupCache: %w", err)
	}

	app.cache = inMemory

	return nil
}

func (app *App) setupEncodedURLStore(_ context.Context) error {
	store, err := infrastructure.NewEncodedURLStore(app.postgresClients.ShortorgClient, app.cache)
	if err != nil {
		return fmt.Errorf("setupEncodedUrlStore: %w", err)
	}
	app.encodedURLStore = store
	return nil
}

func (app *App) setupTokenKeyStore(ctx context.Context) error {
	store, err := infrastructure.NewTokenKeyStore(
		ctx,
		app.postgresClients.TokenIdentifierClient,
		app.logger,
		app.config.TokenKeyStoreBuffer,
	)
	if err != nil {
		return fmt.Errorf("setupTokenKeyStore: %w", err)
	}
	app.tokenKeyStore = store

	return nil
}

func (app *App) setupUseCaseFns(_ context.Context) error { //nolint:unparam // error in the future
	app.urlWasEncodedChan = make(chan encode.UrlWasEncoded, app.config.EncodedUrlsQueSize)
	app.encodeFn = encode.NewEncodeFn(
		app.tokenKeyStore,
		infrastructure.UrlParser{},
		base58.Codec{},
		app.logger,
		app.urlWasEncodedChan,
	)
	app.decodeFn = decode.NewDecodeFn(app.logger, decodeInfrastructure.UrlParser{}, base58.Codec{}, app.encodedURLStore)

	return nil
}

func (app *App) setupEventHandlers(_ context.Context) error { //nolint:unparam // error in the future
	app.urlWasEncodedHandler = encodeInfrastructure.NewUrlWasEncodedHandler(
		app.logger,
		app.encodedURLStore,
		app.config.EncodedUrlsQueSize,
		1,
		app.urlWasEncodedChan,
	)
	return nil
}
