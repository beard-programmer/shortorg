package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/beard-programmer/shortorg/internal/common"
	"github.com/jmoiron/sqlx"
)

type TokenSystemError struct {
	Err error
}

func (e TokenSystemError) Error() string {
	return fmt.Sprintf("Token system error: %v", e.Err)
}

type PostgresIdentifierProvider struct {
	DB *sqlx.DB
}

func (p *PostgresIdentifierProvider) ProduceTokenIdentifier(ctx context.Context) (*common.IntegerBase58Exp5To6, error) {
	ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	var uniqueID int64
	query := "SELECT nextval('token_identifier')"

	err := p.DB.GetContext(ctx, &uniqueID, query)
	if err != nil {
		return nil, TokenSystemError{Err: err}
	}

	return new(common.IntegerBase58Exp5To6).FromInt(uniqueID)
}
