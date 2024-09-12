package encode

import (
	"errors"
	"fmt"
)

type OriginalURL struct {
	url URL
}

func (u OriginalURL) String() string {
	return u.url.String()
}

type URL interface {
	Scheme() string
	Hostname() string
	String() string
}

func OriginalURLFromString(parseUrl UrlParser, s string) (*OriginalURL, error) {
	uri, err := parseUrl.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	if 255 <= len(uri.String()) {
		return nil, errors.New("value is too long! Max 255 characters allowed")
	}

	if uri.Scheme() != "http" && uri.Scheme() != "https" {
		return nil, fmt.Errorf("invalid scheme %s", uri.Scheme())
	}

	return &OriginalURL{url: uri}, nil
}
