package encode

import (
	"errors"
	"fmt"
)

type OriginalURL struct {
	Value string
}

type URL interface {
	Scheme() string
	Hostname() string
	String() string
}

func OriginalURLFromString(parseUrl func(string) (URL, error), s string) (*OriginalURL, error) {
	uri, err := parseUrl(s)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	if 255 <= len(uri.String()) {
		return nil, errors.New("Value is too long! Max 255 characters allowed")
	}

	if uri.Scheme() != "http" && uri.Scheme() != "https" {
		return nil, fmt.Errorf("invalid scheme %s", uri.Scheme())
	}

	return &OriginalURL{Value: uri.String()}, nil
}
