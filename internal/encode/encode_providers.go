package encode

import (
	"context"

	"github.com/beard-programmer/shortorg/internal/core"
)

type TokenKeyStore interface {
	Issue(ctx context.Context) (*core.TokenKey, error)
}

type UrlParser interface {
	core.UrlParser
}

type Encoder interface {
	core.Encoder
}
