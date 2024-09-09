package encode

import (
	"fmt"
)

type RequestValidated struct {
	OriginalURL string
	TokenHost   TokenHost
}

func FromUnvalidatedRequest(parseUrl func(string) (URL, error), url string, encodeAtHost *string) (*RequestValidated, error) {
	encodeWhat, err := new(OriginalURL).FromString(parseUrl, url)
	if err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	encodeWhere, err := TokenHostFromString(encodeAtHost)
	if err != nil {
		return nil, err
	}

	return &RequestValidated{
		OriginalURL: encodeWhat.url,
		TokenHost:   encodeWhere,
	}, nil
}
