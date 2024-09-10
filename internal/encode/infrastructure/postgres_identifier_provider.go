package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/common"
	"github.com/jmoiron/sqlx"
)

type PostgresIdentityProviderError struct {
	Err error
}

func (e PostgresIdentityProviderError) Error() string {
	return fmt.Sprintf("PostgresIdentityProviderError error: %v", e.Err)
}

type PostgresIdentifierProvider struct {
	DB *sqlx.DB
}

func (p *PostgresIdentifierProvider) ProduceTokenIdentifier(ctx context.Context) (*common.IntegerBase58Exp5To6, error) {
	var uniqueID int64
	query := "SELECT nextval('token_identifier')"

	err := p.DB.GetContext(ctx, &uniqueID, query)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, PostgresIdentityProviderError{Err: err}
	}

	return new(common.IntegerBase58Exp5To6).FromInt(uniqueID)
}
