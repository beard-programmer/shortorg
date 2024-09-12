package infrastructure

import (
	"context"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/jmoiron/sqlx"
)

type IdentitiesPostgresError struct {
	Err error
}

func (e IdentitiesPostgresError) Error() string {
	return fmt.Sprintf("IdentitiesPostgres error: %v", e.Err)
}

type IdentitiesPostgres struct {
	DB *sqlx.DB
}

func (p *IdentitiesPostgres) GenerateOne(ctx context.Context) (*encode.Identity, error) {
	tokens, err := p.GenerateMany(ctx, 1)
	if err != nil {
		return nil, err
	}

	return tokens[0], nil
}

func (p *IdentitiesPostgres) GenerateMany(ctx context.Context, bulkSize int) ([]*encode.Identity, error) {
	var uniqueIDs []int64

	err := p.DB.SelectContext(ctx, &uniqueIDs, `SELECT nextval('token_identifier') FROM generate_series(1, $1)`, bulkSize)
	if err != nil {
		return nil, IdentitiesPostgresError{Err: err}
	}

	if len(uniqueIDs) != bulkSize {
		return nil, IdentitiesPostgresError{Err: fmt.Errorf("GenerateMany error: incorrect number of unique ids from DB: %d", len(uniqueIDs))}
	}

	tokens := make([]*encode.Identity, 0, bulkSize)
	for _, id := range uniqueIDs {
		token, err := encode.NewIdentity(id)
		if err != nil {
			return nil, IdentitiesPostgresError{Err: fmt.Errorf("GenerateMany error: failed to convert id %d: %w", id, err)}
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}
