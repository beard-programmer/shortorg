package postgresClients

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/qustavo/sqlhooks/v2"
	"go.uber.org/zap"
)

type Clients struct {
	TokenIdentifierClient *sqlx.DB
	ShortorgClient        *sqlx.DB
}

func New(
	ctx context.Context,
	logger *zap.Logger,
	cfg ClientsConfig,
	appName string,
	isProd bool,
) (*Clients, error) {
	registeredSqlHook := registerSqlHook(logger.Sugar())
	newClientFn := func(ctx context.Context, cfg Config) (*sqlx.DB, error) {
		return newPostgresClient(ctx, logger, cfg, registeredSqlHook, appName, isProd)
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

func newPostgresClient(ctx context.Context, logger *zap.Logger, cfg Config, registeredSqlHook sqlHook, appName string, _ bool) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s application_name=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, appName)

	db, err := sqlx.ConnectContext(ctx, registeredSqlHook.driverName(), connStr)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgtesClient: %w", err)
	}
	logger.Info("Successfully connected to database", zap.String("database", cfg.DBName))
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)

	migrationDB, err := sql.Open(registeredSqlHook.driverName(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection for migrations: %w", err)
	}
	defer migrationDB.Close()

	instance, err := postgres.WithInstance(migrationDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres instance: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://migrations/%s", cfg.DBName), cfg.DBName, instance,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrations: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Migrations applied successfully", zap.String("database", cfg.DBName))

	return db, nil
}

func registerSqlHook(logger *zap.SugaredLogger) sqlHook {
	logger.Info("Registering sql hook")
	hook := sqlHook{logger: logger}
	sql.Register(hook.driverName(), hook.driver())
	return hook
}

type sqlHook struct {
	logger *zap.SugaredLogger
}

func (h *sqlHook) driver() driver.Driver {
	return sqlhooks.Wrap(&pq.Driver{}, h)
}

func (h *sqlHook) driverName() string {
	return "pg-driver-with-sql-hook"
}

type ctxKeyStartTime struct{}

func (h *sqlHook) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
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
		h.logger.Warnln("Sql query took longer than 10ms",
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
		h.logger.Errorw("Query error",
			"query", query,
			"args", args,
			"error", err,
			"duration", duration.Seconds(),
		)
	} else {
		h.logger.Errorw("Query error",
			"query", query,
			"args", args,
			"error", err,
		)
	}

	return err
}
