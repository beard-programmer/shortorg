package encode

import (
	"errors"
)

type Codec interface {
	Encode(int64) string
	Decode(string) (int64, error)
}

func NewToken(codec Codec, tokenIdentifier TokenIdentifier, tokenHost TokenHost) (*TokenStandard, error) {
	switch tokenHost.(type) {
	case *TokenHostStandard:
		return new(TokenStandard).New(codec, tokenIdentifier, tokenHost)
	default:
		return nil, errors.New("only standard tokens are supported")
	}

}
