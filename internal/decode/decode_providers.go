package decode

import (
	"context"

	"github.com/beard-programmer/shortorg/internal/core"
)

type Codec interface {
	core.Encoder
	Decoder
}

type Decoder interface {
	Decode(string) (int64, error)
}

type EncodedUrlsProvider interface {
	FindOne(context.Context, core.TokenKey) (string, error)
}

type EncodedUrlDto interface {
	OriginalUrl() string
}

type UrlParser interface {
	Parse(string) (URL, error)
}

type URL interface {
	core.URL
	Path() string
}

type UrlParserAdapter struct {
	parser UrlParser
}

func (a UrlParserAdapter) Parse(s string) (core.URL, error) {
	return a.parser.Parse(s)
}
