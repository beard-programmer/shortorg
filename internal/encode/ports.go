package encode

import (
	"context"

	"github.com/beard-programmer/shortorg/internal/core"
)

type LinkKeyStore interface {
	Issue(ctx context.Context) (*core.LinkKey, error)
}

type URLParser interface {
	core.URLParser
}
