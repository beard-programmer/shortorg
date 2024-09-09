package internal

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/beard-programmer/shortorg/internal/app"
	"github.com/beard-programmer/shortorg/internal/encode"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type App struct {
	DB         *sqlx.DB
	IdentityDB *sqlx.DB
	Logger     *zap.SugaredLogger
	Server     *http.Server
}

func (a *App) New() *App {
	environment := getEnvWithDefault("GO_ENV", "development")
	numCores := runtime.GOMAXPROCS(0)

	a.Logger = app.InitZapLogger()
	a.Logger.Infow("Initializing app",
		"environment", environment,
		"numCores", numCores,
	)

	driver := app.RegisterSqlLogger(a.Logger)

	mainDB, err := app.ConnectDb("db.json", environment, driver, a.Logger)
	if err != nil {
		a.Logger.Fatalf("Failed to connect to main DB: %v", err)
	}
	a.DB = mainDB

	identityDB, err := app.ConnectDb("identity_db.json", environment, driver, a.Logger)
	if err != nil {
		a.Logger.Fatalf("Failed to connect to identity DB: %v", err)
	}
	a.IdentityDB = identityDB

	// Setup HTTP server with routes
	a.setupServer()

	return a
}

func (a *App) setupServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/encode", encode.ApiHandler(a.IdentityDB))
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)

	loggedMux := app.LoggingMiddleware(a.Logger, mux)

	a.Server = &http.Server{
		Addr:         ":8080",
		Handler:      loggedMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func (a *App) StartServer() error {
	a.Logger.Info("Starting server on :8080...")
	return a.Server.ListenAndServe()
}

func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
