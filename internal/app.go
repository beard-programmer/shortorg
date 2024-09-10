package internal

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/beard-programmer/shortorg/internal/app"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/beard-programmer/shortorg/internal/encode/infrastructure"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type App struct {
	IdentifierProvider      encode.IdentifierProvider
	RedisIdentifierProvider encode.IdentifierProvider
	ParseUrl                func(s string) (encode.URL, error)
	Logger                  *zap.SugaredLogger
	Worker                  *encode.UrlSaveWorker
}

func (a *App) New(ctx context.Context) *App {
	environment := getEnvWithDefault("GO_ENV", "development")

	a.Logger = app.InitZapLogger()
	a.Logger.Infow("Initializing app",
		"environment", environment,
	)

	driver := app.RegisterSqlLogger(a.Logger)
	identityDB, err := app.ConnectDb("identity_db.json", environment, driver, 50, a.Logger)
	if err != nil {
		a.Logger.Fatalf("Failed to connect to identity DB: %v", err)
	}

	redis, err := app.ConnectToRedis(ctx, a.Logger, environment)
	if err != nil {
		a.Logger.Fatalf("Failed to connect to redis: %v", err)
	}

	a.RedisIdentifierProvider = &infrastructure.RedisIdentifierProvider{Redis: redis}
	a.IdentifierProvider = &infrastructure.PostgresIdentifierProvider{DB: identityDB}

	mainDB, err := app.ConnectDb("db.json", environment, driver, 10, a.Logger)
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
	workerErrChan := make(chan error)
	a.Worker.Start(ctx, workerErrChan)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/encode", encode.ApiHandler(a.RedisIdentifierProvider, a.ParseUrl, a.Logger, a.Worker.GetEventChan()))
	mux.HandleFunc(
		"/encode2", encode.ApiHandler(a.IdentifierProvider, a.ParseUrl, a.Logger, a.Worker.GetEventChan()))
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
	case err := <-workerErrChan:
		a.Logger.Errorf("Worker returned error, cant proceed: %v", err)
		return err
	}

}

func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
