package encode

import (
	"errors"
)

func NewToken(codec CodecProvider, identity Identity, tokenHost TokenHost, originalUrl OriginalURL) (*TokenStandard, error) {
	switch tokenHost.(type) {
	case *TokenHostStandard:
		return NewTokenStandard(codec, identity, tokenHost, originalUrl)
	default:
		return nil, errors.New("only standard tokens are supported")
	}

}
