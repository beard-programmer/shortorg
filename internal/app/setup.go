package app

import (
	"context"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/core/infrastructure"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/cache"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/postgresClients"
	"github.com/beard-programmer/shortorg/internal/encode"
)

func (app *App) Setup(ctx context.Context) error {
	//err := app.setupConfig(ctx)
	//if err != nil {
	//	return fmt.Errorf("setup config: %w", err)
	//}
	//
	//err = app.setupContextUtils(ctx)
	//if err != nil {
	//	return fmt.Errorf("setup context utils: %w", err)
	//}

	err := app.setupPostgresClients(ctx)
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

	return nil
}

//
//func (app *App) setupConfig(ctx context.Context) error {
//	var err error
//
//	app.config, err = config.ParseConfig(ctx)
//	if err != nil {
//		return fmt.Errorf("parse config: %w", err)
//	}
//
//	vaultInitCtx, vaultInitCtxCancel := context.WithTimeout(ctx, time.Minute)
//	defer vaultInitCtxCancel()
//
//	vaultClient, err := vault.NewVaultConfigSource(vaultInitCtx, app.logger, app.config.Vault, app.config.ENV)
//	if err != nil {
//		return fmt.Errorf("create vault client: %w", err)
//	}
//
//	err = config.FillFieldsFromSource(ctx, app.config, vaultClient)
//	if err != nil {
//		return fmt.Errorf("fetch config: %w", err)
//	}
//
//	return nil
//}

func (app *App) setupPostgresClients(ctx context.Context) error {
	clients, err := postgresClients.New(ctx, app.logger, app.config.postgresClientsConfig, app.Name(), app.config.IsProdEnv())
	if err != nil {
		return fmt.Errorf("setupPostgresClients: %w", err)
	}

	app.postgresClients = clients
	return nil
}

func (app *App) setupCache(ctx context.Context) error {
	inMemory, err := cache.NewInMemory[string](app.config.cacheConfig)
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
	encodeFn := encode.NewEncodeFn()
	return nil
}

//
//func (app *App) setupContextUtils(_ context.Context) error {
//	app.contextUtils = contextUtils.New()
//
//	return nil
//}
