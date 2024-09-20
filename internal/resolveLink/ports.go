package resolveLink

import (
	"context"
	"errors"

	"github.com/beard-programmer/shortorg/internal/core"
)

var ErrLinkNotFound = errors.New("link not found")

type LinksStore interface {
	FindOneNonBrandedLink(context.Context, core.LinkSlugDto, core.LinkKeyDto, core.LinkHostDto) (*core.LinkDTO, bool, error)
}

type EncodedUrlDto interface {
	OriginalUrl() string
}
