package encode

import "context"

type KeyIssuer interface {
	Issue(ctx context.Context) (*UnclaimedKey, error)
}

type UrlParser interface {
	Parse(string) (URL, error)
}

type Codec interface {
	Encoder
	Decoder
}

type Encoder interface {
	Encode(int64) string
}

type Decoder interface {
	Decode(string) (int64, error)
}
