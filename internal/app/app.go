package app

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/beard-programmer/shortorg/internal/api"
	"github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/beard-programmer/shortorg/internal/infrastructure"
)

type App struct {
	logger               *logger.AppLogger
	cfg                  config
	encodeFn             encode.Fn
	urlWasEncodedHandler encode.SaveEncodedURLJob
	decodeFn             decode.Fn
}

func New(ctx context.Context, logger *logger.AppLogger) (*App, error) {
	env := os.Getenv("APP_ENV")
	cfg, err := config{}.load(env)
	if err != nil {
		return nil, fmt.Errorf("app.ConnectToPostgresClients: setup cfg: %w", err)
	}

	logger.InfoContext(ctx, "app.ConnectToPostgresClients: cfg was set up", "cfg", cfg)

	_ = runtime.GOMAXPROCS(cfg.Concurrency)

	postgresClients, err := infrastructure.ConnectToPostgresClients(
		ctx,
		logger,
		cfg.Infrastructure.PostgresClients,
		Name(),
		cfg.isProdEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("app.ConnectToPostgresClients: setup postgres clients: %w", err)
	}

	encodedURLCache, err := infrastructure.NewCache[string](cfg.Infrastructure.Cache)
	if err != nil {
		return nil, fmt.Errorf("app.ConnectToPostgresClients: setupEncodedURLStore: %w", err)
	}

	encodedURLStore, err := infrastructure.NewEncodedURLStore(postgresClients.ShortorgClient, encodedURLCache, logger)
	if err != nil {
		return nil, fmt.Errorf("app.ConnectToPostgresClients: setupEncodedUrlStore: %w", err)
	}

	tokenStore, err := infrastructure.NewLinkKeyStore(
		ctx,
		logger,
		postgresClients.TokenIdentifierClient,
		cfg.Infrastructure.TokenStore,
	)
	if err != nil {
		return nil, fmt.Errorf("app.ConnectToPostgresClients: setup token key store: %w", err)
	}

	urlWasEncodedChan := make(chan encode.URLWasEncoded, cfg.EncodedUrlsQueSize)
	encodeFn := encode.NewEncodeFn(tokenStore, infrastructure.UrlParser{}, logger, urlWasEncodedChan)
	decodeFn := decode.NewDecodeFn(logger, infrastructure.UrlParser{}, encodedURLStore)

	urlWasEncodedHandler := encode.NewSaveEncodedURLJob(
		logger,
		encodedURLStore,
		cfg.EncodedUrlsQueSize,
		1,
		urlWasEncodedChan,
	)

	return &App{
		logger:               logger,
		cfg:                  *cfg,
		encodeFn:             encodeFn,
		decodeFn:             decodeFn,
		urlWasEncodedHandler: urlWasEncodedHandler,
	}, nil
}

func Name() string {
	return "shortorg"
}

func (app *App) Serve(ctx context.Context) error {
	server := api.New(
		app.encodeFn,
		app.decodeFn,
		app.urlWasEncodedHandler,
		app.logger,
		app.cfg.APIServer,
		Name(),
		app.cfg.Env,
	)

	err := server.Serve(ctx)
	if err != nil {
		return fmt.Errorf("server server: %w", err)
	}

	return nil
}
