package internal

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/beard-programmer/shortorg/internal/app"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/beard-programmer/shortorg/internal/encode/infrastructure"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type App struct {
	IdentifierProvider encode.IdentifierProvider
	ParseUrl           func(s string) (encode.URL, error)
	Logger             *zap.SugaredLogger
	Server             *http.Server
	Worker             *encode.UrlSaveWorker
}

func (a *App) New() *App {
	environment := getEnvWithDefault("GO_ENV", "development")
	concurrency := runtime.GOMAXPROCS(0)

	a.Logger = app.InitZapLogger()
	a.Logger.Infow("Initializing app",
		"environment", environment,
		"concurrency", concurrency,
	)

	driver := app.RegisterSqlLogger(a.Logger)
	identityDB, err := app.ConnectDb("identity_db.json", environment, driver, concurrency, a.Logger)
	if err != nil {
		a.Logger.Fatalf("Failed to connect to identity DB: %v", err)
	}

	a.IdentifierProvider = &infrastructure.PostgresIdentifierProvider{DB: identityDB}

	mainDB, err := app.ConnectDb("db.json", environment, driver, 1, a.Logger)
	if err != nil {
		a.Logger.Fatalf("Failed to connect to main DB: %v", err)
	}
	a.Worker = new(encode.UrlSaveWorker).New(&infrastructure.SaveEncodedUrlProvider{DB: mainDB}, a.Logger)

	a.ParseUrl = func(s string) (encode.URL, error) {
		return infrastructure.ParseURLString(s)
	}

	return a
}

func (a *App) StartServer(ctx context.Context) error {
	a.Worker.Start(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/encode", encode.ApiHandler(a.IdentifierProvider, a.ParseUrl, a.Logger, a.Worker.GetEventChan()))
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)

	loggedMux := app.LoggingMiddleware(a.Logger, mux)

	a.Server = &http.Server{
		Addr:         ":8080",
		Handler:      loggedMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	serverErrChan := make(chan error, 1)

	go func() {
		a.Logger.Info("Starting server on :8080...")
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
		close(serverErrChan)
	}()

	select {
	case <-ctx.Done():
		a.Logger.Info("Context canceled, shutting down server...")

		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
		}()

		if err := a.Server.Shutdown(ctxShutDown); err != nil {
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
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
