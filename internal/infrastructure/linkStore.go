package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/core"
	"github.com/jmoiron/sqlx"
)

var errEncodedURLStore = errors.New("errEncodedURLStore")

type LinkStore struct {
	postgresClient *sqlx.DB
	cache          Cache[string]
	logger         *logger.AppLogger
}

type Cache[T any] interface {
	Get(context.Context, any) (T, error)
	Set(context.Context, any, T) error
}

func NewEncodedURLStore(postgresClient *sqlx.DB, cache Cache[string], logger *logger.AppLogger) (
	*LinkStore,
	error,
) {
	if postgresClient == nil {
		return nil, fmt.Errorf("%w: NewEncodedURLStore: postgresClient is nil", errEncodedURLStore)
	}
	return &LinkStore{postgresClient, cache, logger}, nil
}

func (s *LinkStore) FindOneNonBrandedLink(
	ctx context.Context,
	slugDto core.LinkSlugDto,
	keyDto core.LinkKeyDto,
	hostDto core.LinkHostDto,
) (
	*core.LinkDTO,
	bool,
	error,
) {
	var (
		key  int64
		url  string
		slug string
	)

	row := s.postgresClient.QueryRowxContext(
		ctx,
		"SELECT url, token, token_identifier FROM encoded_urls WHERE token_identifier=$1 LIMIT 1",
		keyDto.Value,
	)

	err := row.Scan(&url, &slug, &key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("%w: FindOne: failed to execute%s", errLinkKeyStore, err)
	}

	return &core.LinkDTO{Key: keyDto, Slug: slugDto, Host: hostDto, DestinationURL: core.URLDto{Value: url}}, true, nil
}

func (s *LinkStore) FindOne(ctx context.Context, key core.LinkKey) (string, bool, error) {
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

func (s *LinkStore) SaveMany(ctx context.Context, links []core.LinkDTO) error {
	// NamedExecContext is generating invalid sql so building query manually.
	valueStrings := make([]string, 0, len(links))
	valueArgs := make([]interface{}, 0, len(links)*3) //nolint:mnd // _

	for i, linkDto := range links {
		valueStrings = append(
			valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3), //nolint:mnd // _
		)
		valueArgs = append(
			valueArgs,
			linkDto.Key.Value,
			linkDto.Slug.Value,
			linkDto.DestinationURL.Value,
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

	//for _, encodedURL := range links {
	//	key := encodedURL.NonBrandedLink.Key.Value()
	//	url := encodedURL.NonBrandedLink.DestinationURL.String()
	//	err = s.cache.Set(ctx, encodedURL.NonBrandedLink.Key.Value(), encodedURL.NonBrandedLink.DestinationURL.String())
	//	if err != nil {
	//		s.logger.WarnContext(ctx, fmt.Sprintf("SaveMany: Error storing in cache key %v value %v", key, url))
	//	}
	//}

	return nil
}
