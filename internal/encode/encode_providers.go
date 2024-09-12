package encode

import (
	"context"

	"github.com/beard-programmer/shortorg/internal/core"
)

type KeyIssuer interface {
	Issue(ctx context.Context) (*UnclaimedKey, error)
}

type UrlParser interface {
	core.UrlParser
}

type Encoder interface {
	Encode(int64) string
}
