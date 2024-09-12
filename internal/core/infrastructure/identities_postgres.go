package infrastructure

import (
	"context"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/core"
	"github.com/jmoiron/sqlx"
)

type TokenKeysPostgresError struct {
	Err error
}

func (e TokenKeysPostgresError) Error() string {
	return fmt.Sprintf("TokenKeysPostgres error: %v", e.Err)
}

type TokenKeysPostgres struct {
	DB *sqlx.DB
}

func (p *TokenKeysPostgres) Issue(ctx context.Context) (*core.TokenKey, error) {
	tokens, err := p.GenerateMany(ctx, 1)
	if err != nil {
		return nil, err
	}

	return tokens[0], nil
}

func (p *TokenKeysPostgres) GenerateMany(ctx context.Context, bulkSize int) ([]*core.TokenKey, error) {
	var uniqueIDs []int64

	err := p.DB.SelectContext(ctx, &uniqueIDs, `SELECT nextval('token_identifier') FROM generate_series(1, $1)`, bulkSize)
	if err != nil {
		return nil, TokenKeysPostgresError{Err: err}
	}

	if len(uniqueIDs) != bulkSize {
		return nil, TokenKeysPostgresError{Err: fmt.Errorf("GenerateMany error: incorrect number of unique ids from DB: %d", len(uniqueIDs))}
	}

	tokens := make([]*core.TokenKey, 0, bulkSize)
	for _, id := range uniqueIDs {
		token, err := core.TokenKey{}.New(id)
		if err != nil {
			return nil, TokenKeysPostgresError{Err: fmt.Errorf("GenerateMany error: failed to convert id %d: %w", id, err)}
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}
