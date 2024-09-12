package encode

import (
	"fmt"
)

type EncodingRequest interface {
	OriginalUrl() string
	Host() *string
}

type ValidatedRequest struct {
	OriginalURL OriginalURL
	TokenHost   TokenHost
}

func NewValidatedRequest(parseUrl func(string) (URL, error), url string, encodeAtHost *string) (*ValidatedRequest, error) {
	encodeWhat, err := OriginalURLFromString(parseUrl, url)
	if err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	encodeWhere, err := TokenHostFromString(encodeAtHost)
	if err != nil {
		return nil, err
	}

	return &ValidatedRequest{
		OriginalURL: *encodeWhat,
		TokenHost:   encodeWhere,
	}, nil
}
