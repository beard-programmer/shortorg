package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/qustavo/sqlhooks/v2"
	_ "github.com/qustavo/sqlhooks/v2"
	"go.uber.org/zap"
)

func ConnectDb(configFile, environment string, driver string, maxConnections int, logger *zap.SugaredLogger) (*sqlx.DB, error) {
	dbConfig, err := NewConfig(configFile, environment)
	if err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s application_name=backend sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)

	db, err := sqlx.Connect(driver, connStr)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxConnections)

	logger.Info("Successfully connected to database.")

	migrationDB, err := sqlx.Open(driver, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection for migrations: %w", err)
	}
	defer migrationDB.Close()

	err = RunMigrations(migrationDB.DB, dbConfig.MigrationsPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB, migrationsPath string, logger *zap.SugaredLogger) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver instance: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath, "postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Migrations applied successfully.")
	return nil
}

type DBConfig struct {
	User           string `json:"user"`
	Password       string `json:"password"`
	Host           string `json:"host"`
	DBName         string `json:"dbname"`
	Port           int    `json:"port"`
	MigrationsPath string `json:"migrationsPath"`
}

func NewConfig(fileName, env string) (*DBConfig, error) {
	filePath := fmt.Sprintf("./config/%s", fileName)

	jsonFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config map[string]DBConfig
	if err := json.Unmarshal(jsonFile, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	envConfig, exists := config[env]
	if !exists {
		return nil, fmt.Errorf("environment %s not found in config", env)
	}

	return &envConfig, nil
}

func RegisterSqlLogger(logger *zap.SugaredLogger) string {
	logger.Infow("Registering pg-hook")
	hook := &Hooks{Logger: logger}
	sql.Register("pg-hooks", sqlhooks.Wrap(&pq.Driver{}, hook))
	return "pg-hooks"
}

type Hooks struct {
	Logger *zap.SugaredLogger
}

type ctxKeyStartTime struct{}

// Before hook runs before executing any query
func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	startTime := time.Now()

	ctx = context.WithValue(ctx, ctxKeyStartTime{}, startTime)
	//h.Logger.Warnf("Query started",
	//	"query", query,
	//	"args", args, // Duration in seconds
	//)

	return ctx, nil
}

func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	// Retrieve the start time from the context
	startTime, ok := ctx.Value(ctxKeyStartTime{}).(time.Time)
	if !ok {
		h.Logger.Error("Failed to retrieve start time from context")
		return ctx, nil
	}

	duration := time.Since(startTime)

	if 200*time.Millisecond < duration {
		h.Logger.Warnf("Query completed",
			"query", query,
			"args", args,
			"duration", duration, // Duration in seconds
		)
	}

	return ctx, nil
}

func (h *Hooks) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
	startTime, ok := ctx.Value(ctxKeyStartTime{}).(time.Time)
	if ok {
		duration := time.Since(startTime)
		h.Logger.Errorw("Query error",
			"query", query,
			"args", args,
			"error", err,
			"duration", duration.Seconds(), // Log duration even on error
		)
	} else {
		h.Logger.Errorw("Query error",
			"query", query,
			"args", args,
			"error", err,
		)
	}

	return err
}
