package encode

import "context"

type Identities interface {
	GenerateOne(ctx context.Context) (*Identity, error)
}

type UrlProvider interface {
	Parse(string) (URL, error)
}

type CodecProvider interface {
	EncodingProvider
	DecodingProvider
}

type EncodingProvider interface {
	Encode(int64) string
}

type DecodingProvider interface {
	Decode(string) (int64, error)
}
