package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/beard-programmer/shortorg/internal/core"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/dgraph-io/ristretto"
	ekoCache "github.com/eko/gocache/lib/v4/cache"
	ristrettoStore "github.com/eko/gocache/store/ristretto/v4"
	"github.com/jmoiron/sqlx"
)

// TODO: fix errors chaos
type EncodedUrlProviderPostgresError struct {
	Err error
}

func (e EncodedUrlProviderPostgresError) Error() string {
	return fmt.Sprintf("EncodedUrlsStorage error: %v", e.Err)
}

type encodedUrlsCache = ekoCache.Cache[string]

func newEncodedUrlsCache(config ristretto.Config) (*encodedUrlsCache, error) {
	ristrettoCache, err := ristretto.NewCache(&config)
	if err != nil {
		return nil, err
	}
	rStore := ristrettoStore.NewRistretto(ristrettoCache)
	cacheManager := ekoCache.New[string](rStore)
	return cacheManager, nil
}

type EncodedUrlsStorage struct {
	db    *sqlx.DB
	cache *encodedUrlsCache
}

func (EncodedUrlsStorage) New(db *sqlx.DB, cacheConfig *ristretto.Config) (*EncodedUrlsStorage, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if cacheConfig == nil {
		cacheConfig = &ristretto.Config{
			NumCounters: 10000000, // number of keys to track frequency of (10M).
			MaxCost:     1 << 30,  // maximum cost of cache (1GB).
			BufferItems: 64,       // number of keys per Get buffer.
		}
	}

	cache, err := newEncodedUrlsCache(*cacheConfig)
	if err != nil {
		return nil, err
	}

	return &EncodedUrlsStorage{db, cache}, nil
}

type EncodedUrl struct {
	TokenIdentifier int64  `db:"token_identifier"`
	Url             string `db:"url"`
}

func (e EncodedUrl) OriginalUrl() string {
	return e.Url
}

func (p *EncodedUrlsStorage) FindOne(ctx context.Context, key core.TokenKey) (string, error) {
	var url string

	url, notFound := p.cache.Get(ctx, key.Value())
	if notFound == nil {
		return url, nil
	}

	row := p.db.QueryRowxContext(ctx, "SELECT url FROM encoded_urls WHERE token_identifier=$1 LIMIT 1", key.Value())
	err := row.Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", EncodedUrlProviderPostgresError{fmt.Errorf("FindOne error: failed to execute: %w", err)}
	}

	err = p.cache.Set(ctx, key.Value(), url)
	if err != nil {
		fmt.Printf("FindOne: Error storing in cache key %v value %v", key, url)
	}
	return url, nil
}

func (p *EncodedUrlsStorage) SaveMany(ctx context.Context, encodedUrls []encode.UrlWasEncoded) error {
	// Note: NamedExecContext is generating invalid sql so building query manually.
	valueStrings := make([]string, 0, len(encodedUrls))
	valueArgs := make([]interface{}, 0, len(encodedUrls)*2)

	for i, encodedUrl := range encodedUrls {
		// Prepare placeholder for each row. $1, $2, $3, $4...
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, encodedUrl.Token.Key.Value(), encodedUrl.Token.OriginalURL.String())
	}

	query := fmt.Sprintf(
		"INSERT INTO encoded_urls (token_identifier, url) VALUES %s",
		strings.Join(valueStrings, ","),
	)

	_, err := p.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return EncodedUrlProviderPostgresError{fmt.Errorf("SaveMany error: failed to execute bulk insert: %w", err)}
	}

	for _, encodedUrl := range encodedUrls {
		key := encodedUrl.Token.Key.Value()
		url := encodedUrl.Token.OriginalURL.String()
		err = p.cache.Set(ctx, encodedUrl.Token.Key.Value(), encodedUrl.Token.OriginalURL.String())
		if err != nil {
			fmt.Printf("SaveMany: Error storing in cache key %v value %v", key, url)
		}
	}

	return nil
}
