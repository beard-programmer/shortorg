package app

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/beard-programmer/shortorg/internal/api"
	"github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/base58"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/cache"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/postgres"
	"github.com/beard-programmer/shortorg/internal/decode"
	decodeInfrastructure "github.com/beard-programmer/shortorg/internal/decode/infrastructure"
	"github.com/beard-programmer/shortorg/internal/encode"
	encodeInfrastructure "github.com/beard-programmer/shortorg/internal/encode/infrastructure"
)

type App struct {
	logger               *logger.AppLogger
	cfg                  config
	encodeFn             encode.Fn
	urlWasEncodedHandler encodeInfrastructure.URLWasEncodedHandlerFn
	decodeFn             decode.Fn
}

func New(ctx context.Context, logger *logger.AppLogger) (*App, error) {
	env := os.Getenv("APP_ENV")
	cfg, err := config{}.load(env)
	if err != nil {
		return nil, fmt.Errorf("app.New: setup cfg: %w", err)
	}

	logger.Sugar().Infow("app.New: cfg was set up", "cfg", cfg)

	_ = runtime.GOMAXPROCS(cfg.Concurrency)

	clients, err := postgres.New(
		ctx,
		logger,
		cfg.PostgresClients,
		Name(),
		cfg.isProdEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("app.New: setup postgres clients: %w", err)
	}

	postgresClients := clients

	encodedUrlCache, err := cache.NewCache[string](cfg.Cache)
	if err != nil {
		return nil, fmt.Errorf("app.New: setupEncodedURLStore: %w", err)
	}

	encodedURLStore, err := infrastructure.NewEncodedURLStore(postgresClients.ShortorgClient, encodedUrlCache, logger)
	if err != nil {
		return nil, fmt.Errorf("app.New: setupEncodedUrlStore: %w", err)
	}

	tokenStore, err := infrastructure.NewTokenKeyStore(
		ctx,
		logger,
		postgresClients.TokenIdentifierClient,
		cfg.Infrastructure.TokenStore,
	)
	if err != nil {
		return nil, fmt.Errorf("app.New: setup token key store: %w", err)
	}

	urlWasEncodedChan := make(chan encode.UrlWasEncoded, cfg.EncodedUrlsQueSize)
	encodeFn := encode.NewEncodeFn(
		tokenStore,
		infrastructure.UrlParser{},
		base58.Codec{},
		logger,
		urlWasEncodedChan,
	)
	decodeFn := decode.NewDecodeFn(logger, decodeInfrastructure.UrlParser{}, base58.Codec{}, encodedURLStore)

	urlWasEncodedHandler := encodeInfrastructure.NewUrlWasEncodedHandler(
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
	)

	err := server.Serve(ctx)
	if err != nil {
		return fmt.Errorf("server server: %w", err)
	}

	return nil
}
