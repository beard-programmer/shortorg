package encode

import (
	"errors"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/base58"
	"github.com/beard-programmer/shortorg/internal/core"
)

type UnclaimedKey = base58.IntegerExp5To6
type TokenHost interface {
	Hostname() string
}

func TokenHostFromString(host *string) (TokenHost, error) {
	if host == nil || *host == "" || *host == StandardTokenHost {
		return &TokenHostStandard{}, nil
	}

	return nil, fmt.Errorf("not allowed to encode URL using host %s", *host)
}

const StandardTokenHost = "shortl.org"

type TokenHostStandard struct{}

func (t *TokenHostStandard) Hostname() string {
	return StandardTokenHost
}

type TokenKeyEncoded = base58.StringExp5To6

type TokenStandard struct {
	Key         UnclaimedKey
	KeyEncoded  TokenKeyEncoded
	Host        TokenHost
	OriginalURL core.OriginalURL
}

func NewToken(codec Encoder, tokenKey UnclaimedKey, tokenHost TokenHost, originalUrl core.OriginalURL) (*TokenStandard, error) {
	switch tokenHost.(type) {
	case *TokenHostStandard:
		return TokenStandard{}.new(codec, tokenKey, tokenHost, originalUrl)
	default:
		return nil, errors.New("only standard tokens are supported")
	}
}

func (TokenStandard) new(codec Encoder, tokenKey UnclaimedKey, tokenHost TokenHost, originalUrl core.OriginalURL) (*TokenStandard, error) {
	tokenKeyEncoded, err := TokenKeyEncoded{}.New(codec.Encode(tokenKey.Value()))
	if err != nil {
		return nil, err
	}

	return &TokenStandard{tokenKey, *tokenKeyEncoded, tokenHost, originalUrl}, nil
}
