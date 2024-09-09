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

//
//type DBConfig struct {
//	User     string `json:"user"`
//	Password string `json:"password"`
//	Host     string `json:"host"`
//	DBName   string `json:"dbname"`
//	Port     int    `json:"port"`
//}
//
//func (c *DBConfig) New(fileName, env string) (*DBConfig, error) {
//	filePath := fmt.Sprintf("./config/%s", fileName)
//
//	// Use os.ReadFile to read the file
//	jsonFile, err := os.ReadFile(filePath)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read config file: %v", err)
//	}
//
//	// Parse the JSON into a map of environments
//	var config map[string]DBConfig
//	if err := json.Unmarshal(jsonFile, &config); err != nil {
//		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
//	}
//
//	envConfig, exists := config[env]
//	if !exists {
//		return nil, fmt.Errorf("environment %s not found in config", env)
//	}
//
//	return &envConfig, nil
//}
//
//type Hooks struct {
//	Logger *zap.SugaredLogger
//}
//
//type ctxKeyStartTime struct{}
//
//// Before hook runs before executing any query
//func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
//	startTime := time.Now()
//
//	ctx = context.WithValue(ctx, ctxKeyStartTime{}, startTime)
//
//	return ctx, nil
//}
//
//func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
//	// Retrieve the start time from the context
//	startTime, ok := ctx.Value(ctxKeyStartTime{}).(time.Time)
//	if !ok {
//		h.Logger.Error("Failed to retrieve start time from context")
//		return ctx, nil
//	}
//
//	duration := time.Since(startTime)
//
//	if 100*time.Millisecond < duration {
//		h.Logger.Warnf("Query completed",
//			"query", query,
//			"args", args,
//			"duration", duration, // Duration in seconds
//		)
//	}
//
//	return ctx, nil
//}
//
//func (h *Hooks) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
//	startTime, ok := ctx.Value(ctxKeyStartTime{}).(time.Time)
//	if ok {
//		duration := time.Since(startTime)
//		h.Logger.Errorw("Query error",
//			"query", query,
//			"args", args,
//			"error", err,
//			"duration", duration.Seconds(), // Log duration even on error
//		)
//	} else {
//		h.Logger.Errorw("Query error",
//			"query", query,
//			"args", args,
//			"error", err,
//		)
//	}
//
//	return err
//}
//
//type App struct {
//	DB         *sqlx.DB
//	IdentityDB *sqlx.DB
//	Logger     *zap.SugaredLogger
//	Server     *http.Server
//}
//
//func InitZapLogger() *zap.SugaredLogger {
//	config := zap.NewProductionEncoderConfig()
//
//	config.EncodeTime = zapcore.ISO8601TimeEncoder
//	config.EncodeLevel = zapcore.CapitalLevelEncoder // INFO, DEBUG in uppercase
//
//	consoleEncoder := zapcore.NewConsoleEncoder(config)
//
//	// Write logs to standard output with the console encoder
//	core := zapcore.NewCore(consoleEncoder, zapcore.Lock(zapcore.AddSync(os.Stdout)), zapcore.InfoLevel)
//
//	// Build the logger
//	logger := zap.New(core, zap.AddCaller())
//	return logger.Sugar()
//}
//
//func (a *App) New() *App {
//	environment := getEnvWithDefault("GO_ENV", "development")
//	numCores := runtime.GOMAXPROCS(0)
//
//	a.Logger = InitZapLogger()
//	a.Logger.Infow("Initializing app",
//		"environment", environment,
//		"numCores", numCores,
//	)
//	hook := &Hooks{Logger: a.Logger}
//	sql.Register("pg-hooks", sqlhooks.Wrap(&pq.Driver{}, hook))
//
//	dbConfig, err := new(DBConfig).New("db.json", environment)
//	if err != nil {
//		log.Fatalf("Error loading db config: %v", err)
//	}
//
//	identityConfig, err := new(DBConfig).New("identity_db.json", environment)
//	if err != nil {
//		log.Fatalf("Error loading identity db config: %v", err)
//	}
//	mainDBConnectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
//		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)
//
//	mainDB, err := sqlx.ConnectDb("pg-hooks", mainDBConnectionString)
//	if err != nil {
//		log.Fatalf("Failed to connect to main DB: %v", err)
//	}
//	mainDB.SetMaxOpenConns(100)
//	a.DB = mainDB
//	a.Logger.Info("Successfully connected to main database.")
//	err = a.migrateDatabase(a.DB.DB, "./migrations/db/")
//	if err != nil {
//		log.Fatalf("failed to migrate main DB: %v", err)
//	}
//
//	identityDBConnectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
//		identityConfig.Host, identityConfig.Port, identityConfig.User, identityConfig.Password, identityConfig.DBName)
//
//	identityDB, err := sqlx.ConnectDb("pg-hooks", identityDBConnectionString)
//	if err != nil {
//		log.Fatalf("Failed to connect to identity DB: %v", err)
//	}
//	identityDB.SetMaxOpenConns(100)
//	a.IdentityDB = identityDB
//	a.Logger.Info("Successfully connected to identity database.")
//	err = a.migrateDatabase(a.IdentityDB.DB, "./migrations/identity_db/")
//	if err != nil {
//		log.Fatalf("failed to migrate Identity DB: %v", err)
//	}
//
//	a.setupServer()
//	a.Logger.Info("Server routes successfully set up.")
//
//	return a
//}
//
//func (a *App) setupServer() {
//	mux := http.NewServeMux()
//	mux.HandleFunc("/encode", encode.ApiHandler(a.IdentityDB))
//	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
//	// Wrap the mux with the logging middleware
//	loggedMux := LoggingMiddleware(a.Logger, mux)
//
//	a.Server = &http.Server{
//		Addr:         ":8080",
//		Handler:      loggedMux, // Use the loggedMux with the middleware
//		ReadTimeout:  5 * time.Second,
//		WriteTimeout: 5 * time.Second,
//		IdleTimeout:  10 * time.Second,
//	}
//}
//
//func (a *App) StartServer() error {
//	a.Logger.Info("Starting server on :8080...")
//	return a.Server.ListenAndServe()
//}
//
//func (a *App) migrateDatabase(db *sql.DB, migrationsPath string) error {
//	// Create a new instance of the postgres driver
//	driver, err := postgres.WithInstance(db, &postgres.Config{})
//	if err != nil {
//		return fmt.Errorf("failed to create postgres driver instance: %w", err)
//	}
//
//	// Initialize the migration instance
//	m, err := migrate.NewWithDatabaseInstance(
//		"file://"+migrationsPath, // Path to the migration files
//		"postgres",               // The name of the database
//		driver,
//	)
//	if err != nil {
//		return fmt.Errorf("failed to initialize migrations: %w", err)
//	}
//
//	// Run up migration to apply any pending migrations
//	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
//		return fmt.Errorf("failed to run migrations: %w", err)
//	}
//
//	a.Logger.Info("Migrations applied successfully.")
//	return nil
//}

//
//func LoggingMiddleware(logger *zap.SugaredLogger, next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		start := time.Now()
//
//		// Wrap the response writer to capture the status code and response size
//		wrappedWriter := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
//
//		// Process the request
//		next.ServeHTTP(wrappedWriter, r)
//
//		// Calculate the duration the request took
//		duration := time.Since(start)
//
//		if 1000*time.Millisecond < duration {
//			logger.Warnf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" %.3fms",
//				r.RemoteAddr,
//				time.Now().Format("02/Jan/2006:15:04:05 -0700"),
//				r.Method,
//				r.URL.String(),
//				r.Proto,
//				wrappedWriter.statusCode,
//				wrappedWriter.responseSize,
//				r.UserAgent(),
//				duration,
//			)
//		}
//	})
//}
//
//type statusRecorder struct {
//	http.ResponseWriter
//	statusCode   int
//	responseSize int
//}
//
//func (r *statusRecorder) WriteHeader(code int) {
//	r.statusCode = code
//	r.ResponseWriter.WriteHeader(code)
//}
//func (r *statusRecorder) Write(b []byte) (int, error) {
//	size, err := r.ResponseWriter.Write(b)
//	r.responseSize += size
//	return size, err
//}
