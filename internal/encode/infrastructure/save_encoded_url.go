package infrastructure

import "github.com/jmoiron/sqlx"

func SaveEncodedURL(db *sqlx.DB, originalURL string, token_identifier int) error {
	_, err := db.Exec("INSERT INTO encoded_urls (original_url, token_identifier) VALUES ($1, $2)", originalURL, token_identifier)
	return err
}
