package core

import (
	"errors"

	"github.com/beard-programmer/shortorg/internal/base58"
)

type TokenKey struct {
	base58.IntegerExp5To6
}

func (TokenKey) New(v int64) (*TokenKey, error) {
	key, err := base58.IntegerExp5To6{}.New(v)
	if err != nil {
		return nil, err
	}
	return &TokenKey{IntegerExp5To6: *key}, nil
}

func (k TokenKey) Encode(encoder Encoder) (*TokenKeyEncoded, error) {
	encoded := encoder.Encode(k.Value())
	return TokenKeyEncoded{}.New(encoded)
}

type TokenKeyEncoded struct {
	base58.StringExp5To6
}

func (TokenKeyEncoded) New(s string) (*TokenKeyEncoded, error) {
	encoded, err := base58.StringExp5To6{}.New(s)
	if err != nil {
		return nil, err
	}
	return &TokenKeyEncoded{StringExp5To6: *encoded}, nil
}

func (k TokenKeyEncoded) Decode(decoder Decoder) (*TokenKey, error) {
	decodedKey, err := decoder.Decode(k.Value())
	if err != nil {
		return nil, err
	}

	return TokenKey{}.New(decodedKey)
}

type TokenStandard struct {
	Key         TokenKey
	KeyEncoded  TokenKeyEncoded
	Host        TokenHost
	OriginalURL OriginalURL
}

func NewToken(codec Encoder, tokenKey TokenKey, tokenHost TokenHost, originalUrl OriginalURL) (*TokenStandard, error) {
	switch tokenHost.(type) {
	case *tokenHostStandard:
		return TokenStandard{}.new(codec, tokenKey, tokenHost, originalUrl)
	default:
		return nil, errors.New("only standard tokens are supported")
	}
}

func (TokenStandard) new(encoder Encoder, tokenKey TokenKey, tokenHost TokenHost, originalUrl OriginalURL) (*TokenStandard, error) {
	tokenKeyEncoded, err := tokenKey.Encode(encoder)
	if err != nil {
		return nil, err
	}

	return &TokenStandard{tokenKey, *tokenKeyEncoded, tokenHost, originalUrl}, nil
}
