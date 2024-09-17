package app

import (
	"context"
	"fmt"

	apiServer "github.com/beard-programmer/shortorg/internal/api"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/postgres"
	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
	encodeInfrastructure "github.com/beard-programmer/shortorg/internal/encode/infrastructure"
	"go.uber.org/zap"
)

type App struct {
	logger               *zap.Logger
	postgresClients      *postgres.Clients
	cache                infrastructure.Cache
	encodedURLStore      *infrastructure.EncodedURLStore
	tokenKeyStore        *infrastructure.TokenKeyStore
	urlWasEncodedChan    chan encode.UrlWasEncoded
	encodeFn             encode.Fn
	decodeFn             decode.Fn
	urlWasEncodedHandler encodeInfrastructure.URLWasEncodedHandlerFn
	config               config
}

func New(ctx context.Context, l *zap.Logger) (*App, error) {
	app := App{logger: l} //nolint:exhaustruct // setup constructor
	if err := app.setup(ctx); err != nil {
		return nil, err
	}

	return &app, nil
}

func (*App) Name() string {
	return "shortorg"
}

func (app *App) Serve(ctx context.Context) error {
	api := apiServer.New(
		app.encodeFn,
		app.decodeFn,
		app.urlWasEncodedHandler,
		app.logger,
		app.config.APIServer,
		app.Name(),
	)

	err := api.Serve(ctx)
	if err != nil {
		return fmt.Errorf("api server: %w", err)
	}

	return nil
}
