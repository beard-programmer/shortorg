package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/beard-programmer/shortorg/internal/core"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/jmoiron/sqlx"
)

// TODO: fix errors chaos
type encodedURLProviderPostgresError struct {
	Err error
}

func (e encodedURLProviderPostgresError) Error() string {
	return fmt.Sprintf("EncodedUrlStore error: %v", e.Err)
}

type EncodedURLStore struct {
	postgresClient *sqlx.DB
	cache          Cache
}

type Cache interface {
	Get(context.Context, any) (string, error)
	Set(context.Context, any, string) error
}

func NewEncodedURLStore(db *sqlx.DB, c Cache) (*EncodedURLStore, error) {
	if db == nil {
		return nil, errors.New("postgresClient is nil")
	}
	return &EncodedURLStore{db, c}, nil
}

func (s *EncodedURLStore) FindOne(ctx context.Context, key core.TokenKey) (string, bool, error) {
	var (
		url string
		err error
	)

	url, err = s.cache.Get(ctx, key.Value())
	if nil == err {
		return url, true, err
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
		return url, false, encodedURLProviderPostgresError{fmt.Errorf("FindOne error: failed to execute: %w", err)}
	}

	setCacheErr := s.cache.Set(ctx, key.Value(), url)
	if setCacheErr != nil {
		fmt.Printf("FindOne: Error storing in cache key %v value %v", key, url)
	}
	return url, true, err
}

func (s *EncodedURLStore) SaveMany(ctx context.Context, encodedUrls []encode.UrlWasEncoded) error {
	// NamedExecContext is generating invalid sql so building query manually.
	valueStrings := make([]string, 0, len(encodedUrls))
	valueArgs := make([]interface{}, 0, len(encodedUrls)*2)

	for i, encodedUrl := range encodedUrls {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, encodedUrl.Token.Key.Value(), encodedUrl.Token.OriginalURL.String())
	}

	query := fmt.Sprintf(
		"INSERT INTO encoded_urls (token_identifier, url) VALUES %s",
		strings.Join(valueStrings, ","),
	)

	_, err := s.postgresClient.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return encodedURLProviderPostgresError{fmt.Errorf("SaveMany error: failed to execute bulk insert: %w", err)}
	}

	for _, encodedUrl := range encodedUrls {
		key := encodedUrl.Token.Key.Value()
		url := encodedUrl.Token.OriginalURL.String()
		err = s.cache.Set(ctx, encodedUrl.Token.Key.Value(), encodedUrl.Token.OriginalURL.String())
		if err != nil {
			fmt.Printf("SaveMany: Error storing in cache key %v value %v", key, url)
		}
	}

	return nil
}
