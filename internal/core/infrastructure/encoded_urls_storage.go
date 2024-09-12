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

type EncodedUrlProviderPostgresError struct {
	Err error
}

func (e EncodedUrlProviderPostgresError) Error() string {
	return fmt.Sprintf("EncodedUrlsStorage error: %v", e.Err)
}

type EncodedUrlsStorage struct {
	DB *sqlx.DB
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

	row := p.DB.QueryRowxContext(ctx, "SELECT url FROM encoded_urls WHERE token_identifier=$1 LIMIT 1", key.Value())
	err := row.Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", EncodedUrlProviderPostgresError{fmt.Errorf("FindOne error: failed to execute: %w", err)}
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

	_, err := p.DB.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return EncodedUrlProviderPostgresError{fmt.Errorf("SaveMany error: failed to execute bulk insert: %w", err)}
	}

	return nil
}
