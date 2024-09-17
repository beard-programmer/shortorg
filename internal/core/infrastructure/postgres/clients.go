package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" //
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/qustavo/sqlhooks/v2"
)

type Clients struct {
	TokenIdentifierClient *sqlx.DB
	ShortorgClient        *sqlx.DB
}

func New(
	ctx context.Context,
	logger *appLogger.AppLogger,
	cfg ClientsConfig,
	appName string,
	isProd bool,
) (*Clients, error) {
	registeredSQLHook := registerSQLHook(logger)
	newClientFn := func(ctx context.Context, cfg config) (*sqlx.DB, error) {
		return newPostgresClient(ctx, logger, cfg, registeredSQLHook, appName, isProd)
	}
	shortOrgClient, err := newClientFn(ctx, cfg.ShortOrg)
	if err != nil {
		return nil, fmt.Errorf("error creating shortorg postgres client: %w", err)
	}

	tokenIdentityClient, err := newClientFn(ctx, cfg.TokenIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error creating token identity client: %w", err)
	}

	return &Clients{tokenIdentityClient, shortOrgClient}, nil
}

func newPostgresClient(
	ctx context.Context,
	logger *appLogger.AppLogger,
	cfg config,
	registeredSQLHook sqlHook,
	appName string,
	_ bool,
) (*sqlx.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s application_name=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, appName,
	)

	connection, err := sqlx.ConnectContext(ctx, registeredSQLHook.driverName(), connStr)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgtesClient: %w", err)
	}
	logger.Info("Successfully connected to database", appLogger.String("database", cfg.DBName))
	connection.SetMaxOpenConns(cfg.MaxConnections)
	connection.SetMaxIdleConns(cfg.MaxIdleConnections)

	migrationDB, err := sql.Open(registeredSQLHook.driverName(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection for migrations: %w", err)
	}
	defer func(migrationDB *sql.DB) {
		_ = migrationDB.Close()
	}(migrationDB)

	instance, err := postgres.WithInstance(migrationDB, &postgres.Config{}) //nolint:exhaustruct
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres instance: %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://migrations/%s", cfg.DBName), cfg.DBName, instance,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrations: %w", err)
	}

	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Migrations applied successfully", appLogger.String("database", cfg.DBName))

	return connection, nil
}

func registerSQLHook(logger *appLogger.AppLogger) sqlHook {
	logger.Info("Registering sql hook")
	hook := sqlHook{logger: logger.Sugar()}
	sql.Register(hook.driverName(), hook.driver())

	return hook
}

type sqlHook struct {
	logger *appLogger.SugaredLogger
}

func (h *sqlHook) driver() driver.Driver {
	return sqlhooks.Wrap(&pq.Driver{}, h)
}

func (h *sqlHook) driverName() string {
	return "pg-driver-with-sql-hook"
}

type ctxKeyStartTime struct{}

func (h *sqlHook) Before(ctx context.Context, _ string, _ ...interface{}) (context.Context, error) {
	startTime := time.Now()
	ctx = context.WithValue(ctx, ctxKeyStartTime{}, startTime)

	return ctx, nil
}

func (h *sqlHook) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	startTime, ok := ctx.Value(ctxKeyStartTime{}).(time.Time)
	if !ok {
		h.logger.Error("Failed to retrieve start time from context")

		return ctx, nil
	}

	duration := time.Since(startTime)

	if 100*time.Millisecond < duration {
		h.logger.Warnln(
			"Sql query took longer than 10ms",
			"query", query,
			"args", args,
			"duration", duration,
		)
	}

	return ctx, nil
}

func (h *sqlHook) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
	startTime, ok := ctx.Value(ctxKeyStartTime{}).(time.Time)
	if ok {
		duration := time.Since(startTime)
		h.logger.Errorw(
			"Query error",
			"query", query,
			"args", args,
			"error", err,
			"duration", duration.Seconds(),
		)
	} else {
		h.logger.Errorw(
			"Query error",
			"query", query,
			"args", args,
			"error", err,
		)
	}

	return err
}
