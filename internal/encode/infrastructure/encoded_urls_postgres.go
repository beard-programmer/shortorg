package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/jmoiron/sqlx"
)

type EncodedUrlProviderPostgresError struct {
	Err error
}

func (e EncodedUrlProviderPostgresError) Error() string {
	return fmt.Sprintf("EncodedUrlsPostgres error: %v", e.Err)
}

type EncodedUrlsPostgres struct {
	DB *sqlx.DB
}

func (p *EncodedUrlsPostgres) SaveMany(ctx context.Context, encodedUrls []encode.EncodedUrl) error {
	// Note: NamedExecContext is generating invalid sql so building query manually.
	valueStrings := make([]string, 0, len(encodedUrls))
	valueArgs := make([]interface{}, 0, len(encodedUrls)*2)

	for i, encodedUrl := range encodedUrls {
		// Prepare placeholder for each row. $1, $2, $3, $4...
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, encodedUrl.Token.Identity.Value(), encodedUrl.URL.Value)
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
