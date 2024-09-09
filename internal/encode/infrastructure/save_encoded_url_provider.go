package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/jmoiron/sqlx"
)

type DatabaseError struct {
	Err error
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("Database error: %v", e.Err)
}

type SaveEncodedUrlProvider struct {
	DB *sqlx.DB
}

func (p *SaveEncodedUrlProvider) SaveEncodedURL(ctx context.Context, urls []encode.EncodedUrl) error {
	// Build the query dynamically
	query := "INSERT INTO encoded_urls (url, token_identifier) VALUES "

	// Use placeholders for the values (PostgreSQL uses $1, $2, etc.)
	values := []interface{}{}
	placeholders := []string{}

	for i, url := range urls {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		values = append(values, url.Url, url.TokenIdentifier.Value())
	}

	query += strings.Join(placeholders, ", ")

	// Execute the query
	_, err := p.DB.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to execute bulk insert: %w", err)
	}

	return nil
}
