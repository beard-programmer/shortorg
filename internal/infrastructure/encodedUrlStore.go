package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/core"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/jmoiron/sqlx"
)

var errEncodedURLStore = errors.New("errEncodedURLStore")

type EncodedURLStore struct {
	postgresClient *sqlx.DB
	cache          Cache[string]
	logger         *logger.AppLogger
}

type Cache[T any] interface {
	Get(context.Context, any) (T, error)
	Set(context.Context, any, T) error
}

func NewEncodedURLStore(postgresClient *sqlx.DB, cache Cache[string], logger *logger.AppLogger) (
	*EncodedURLStore,
	error,
) {
	if postgresClient == nil {
		return nil, fmt.Errorf("%w: NewEncodedURLStore: postgresClient is nil", errEncodedURLStore)
	}
	return &EncodedURLStore{postgresClient, cache, logger}, nil
}

func (s *EncodedURLStore) FindOne(ctx context.Context, key core.LinkKey) (string, bool, error) {
	var (
		url string
		err error
	)

	url, err = s.cache.Get(ctx, key.Value())
	if nil == err {
		return url, true, nil
	}

	row := s.postgresClient.QueryRowxContext(
		ctx,
		"SELECT url FROM encoded_urls WHERE token_identifier=$1 LIMIT 1",
		key.Value(),
	)

	err = row.Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return url, false, nil
	}
	if err != nil {
		return url, false, fmt.Errorf("%w: FindOne: failed to execute%s", errLinkKeyStore, err)
	}

	setCacheErr := s.cache.Set(ctx, key.Value(), url)
	if setCacheErr != nil {
		s.logger.WarnContext(ctx, fmt.Sprintf("FindOne: Error storing in cache key %v value %v", key, url))
	}
	return url, true, nil
}

func (s *EncodedURLStore) SaveMany(ctx context.Context, encodedURLs []encode.UrlWasEncoded) error {
	// NamedExecContext is generating invalid sql so building query manually.
	valueStrings := make([]string, 0, len(encodedURLs))
	valueArgs := make([]interface{}, 0, len(encodedURLs)*3) //nolint:mnd // _

	for i, encodedURL := range encodedURLs {
		valueStrings = append(
			valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3), //nolint:mnd // _
		)
		valueArgs = append(
			valueArgs,
			encodedURL.Token.Key.Value(),
			encodedURL.Token.Slug.Value(),
			encodedURL.Token.OriginalURL.String(),
		)
	}

	query := fmt.Sprintf(
		"INSERT INTO encoded_urls (token_identifier, token, url) VALUES %s",
		strings.Join(valueStrings, ","),
	)

	_, err := s.postgresClient.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("%w: SaveMany: failed to execute bulk insert: %s", errLinkKeyStore, err)
	}

	for _, encodedURL := range encodedURLs {
		key := encodedURL.Token.Key.Value()
		url := encodedURL.Token.OriginalURL.String()
		err = s.cache.Set(ctx, encodedURL.Token.Key.Value(), encodedURL.Token.OriginalURL.String())
		if err != nil {
			s.logger.WarnContext(ctx, fmt.Sprintf("SaveMany: Error storing in cache key %v value %v", key, url))
		}
	}

	return nil
}
