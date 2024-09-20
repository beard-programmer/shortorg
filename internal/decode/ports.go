package decode

import (
	"context"

	"github.com/beard-programmer/shortorg/internal/core"
)

type EncodedUrlsProvider interface {
	FindOne(context.Context, core.LinkKey) (string, bool, error)
}

type EncodedUrlDto interface {
	OriginalUrl() string
}
