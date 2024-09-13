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
type EncodedUrlProviderPostgresError struct {
	Err error
}

func (e EncodedUrlProviderPostgresError) Error() string {
	return fmt.Sprintf("EncodedUrlStore error: %v", e.Err)
}

type EncodedUrlStore struct {
	postgresClient *sqlx.DB
	cache          cache
}

type cache interface {
	Get(context.Context, any) (string, error)
	Set(context.Context, any, string) error
}

func NewEncodedUrlStore(db *sqlx.DB, c cache) (*EncodedUrlStore, error) {
	if db == nil {
		return nil, errors.New("postgresClient is nil")
	}
	return &EncodedUrlStore{db, c}, nil
}

//
//func (EncodedUrlStore) New(postgresClient *sqlx.postgresClient, cacheConfig *ristretto.Config) (*EncodedUrlStore, error) {
//	if postgresClient == nil {
//		return nil, errors.New("postgresClient is nil")
//	}
//
//	if cacheConfig == nil {
//		cacheConfig = &ristretto.Config{
//			NumCounters: 100000, // number of keys to track frequency of (100k).
//			MaxCost:     1 << 30,
//			BufferItems: 64, // number of keys per Get buffer.
//		}
//	}
//
//	cache, err := newEncodedUrlsCache(*cacheConfig)
//	if err != nil {
//		return nil, err
//	}
//
//	return &EncodedUrlStore{postgresClient, cache}, nil
//}

type EncodedUrl struct {
	TokenIdentifier int64  `db:"token_identifier"`
	Url             string `db:"url"`
}

func (e EncodedUrl) OriginalUrl() string {
	return e.Url
}

func (s *EncodedUrlStore) FindOne(ctx context.Context, key core.TokenKey) (string, error) {
	var url string

	url, notFound := s.cache.Get(ctx, key.Value())
	if notFound == nil {
		return url, nil
	}

	row := s.postgresClient.QueryRowxContext(ctx, "SELECT url FROM encoded_urls WHERE token_identifier=$1 LIMIT 1", key.Value())
	err := row.Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", EncodedUrlProviderPostgresError{fmt.Errorf("FindOne error: failed to execute: %w", err)}
	}

	err = s.cache.Set(ctx, key.Value(), url)
	if err != nil {
		fmt.Printf("FindOne: Error storing in cache key %v value %v", key, url)
	}
	return url, nil
}

func (s *EncodedUrlStore) SaveMany(ctx context.Context, encodedUrls []encode.UrlWasEncoded) error {
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
		return EncodedUrlProviderPostgresError{fmt.Errorf("SaveMany error: failed to execute bulk insert: %w", err)}
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
