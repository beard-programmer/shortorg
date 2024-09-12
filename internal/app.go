package internal

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/beard-programmer/shortorg/internal/app"
	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
	encodeInfrastructure "github.com/beard-programmer/shortorg/internal/encode/infrastructure"
	infrastructure "github.com/beard-programmer/shortorg/internal/infrastructure"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type App struct {
	identityDb *sqlx.DB
	mainDb     *sqlx.DB
	Logger     *zap.SugaredLogger
}

func (a *App) New(ctx context.Context) *App {
	environment := getEnvWithDefault("GO_ENV", "development")

	a.Logger = app.NewZapLogger()
	a.Logger.Infow("Initializing app",
		"environment", environment,
	)

	driver := app.RegisterSqlLogger(a.Logger)
	identityDB, err := app.ConnectDb(ctx, "identity_db.json", environment, driver, 4, a.Logger)
	if err != nil {
		a.Logger.Fatalf("Failed to connect to identity DB: %v", err)
	}
	a.identityDb = identityDB

	mainDB, err := app.ConnectDb(ctx, "db.json", environment, driver, 20, a.Logger)
	if err != nil {
		a.Logger.Fatalf("Failed to connect to main DB: %v", err)
	}
	a.mainDb = mainDB

	return a
}

func (a *App) StartServer(ctx context.Context) error {
	//const bufferSize = 60 * 1000 // Target RPS
	const bufferSize = 1 // Target RPS

	saveEncodedUrls := infrastructure.ProcessChan(
		a.Logger,
		func(ctx context.Context, encodedUrls []encode.UrlWasEncoded) error {
			provider := encodeInfrastructure.EncodedUrlsPostgres{DB: a.mainDb}
			return provider.SaveMany(ctx, encodedUrls)
		},
	)

	identitiesBuffered, tokenIdentityProviderErrChan := encodeInfrastructure.NewIdentityProviderWithBuffer(ctx, &encodeInfrastructure.IdentitiesPostgres{DB: a.identityDb}, a.Logger, bufferSize)

	urlWasEncodedChan := make(chan encode.UrlWasEncoded, bufferSize)
	encodeUrl := encode.NewEncodeFunc(identitiesBuffered, encodeInfrastructure.UrlParser{}, encodeInfrastructure.CodecBase58{}, a.Logger, urlWasEncodedChan)
	saveEncodedUrlsErrChan := saveEncodedUrls(ctx, bufferSize, 1, 250*time.Millisecond, urlWasEncodedChan)

	decodeUrl := decode.NewDecodeFunc(a.Logger)

	mux := http.NewServeMux()

	mux.Handle("/encode", encode.HttpHandler(a.Logger, encodeUrl))
	mux.Handle("/decode", decode.HttpHandler(a.Logger, decodeUrl))
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)

	loggedMux := app.LoggingMiddleware(a.Logger, mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      loggedMux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrChan := make(chan error, 1)

	go func() {
		a.Logger.Info("Starting server on :8080...")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
		close(serverErrChan)
	}()

	go func() {
		for err := range saveEncodedUrlsErrChan {
			// TODO: do something
			a.Logger.Errorf("save url returned error: %v", err)
		}

	}()

	select {
	case <-ctx.Done():
		a.Logger.Info("Context canceled, shutting down server...")

		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
		}()

		if err := server.Shutdown(ctxShutDown); err != nil {
			a.Logger.Errorf("Server shutdown failed: %v", err)
			return err
		}

		a.Logger.Info("Server successfully shutdown")
		return nil
	case err := <-serverErrChan:
		if err != nil {
			a.Logger.Errorf("Server encountered an error: %v", err)
			return err
		}
		return nil
	//case err := <-saveEncodedUrlsErrChan:
	//	a.Logger.Errorf("save url returned error: %v", err)
	//	return err
	case err := <-tokenIdentityProviderErrChan:
		a.Logger.Errorf("Token producer returned err, cant proceed: %v", err)
		return err
	}

}

func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
