package core

import (
	"errors"
	"fmt"
)

type DestinationURL struct {
	url URL
}

func (u DestinationURL) Scheme() string {
	return u.url.Scheme()
}

func (u DestinationURL) Hostname() string {
	return u.url.Hostname()
}

func (u DestinationURL) String() string {
	return u.url.String()
}

func DestinationURLFromString(parseUrl URLParser, s string) (*DestinationURL, error) {
	uri, err := parseUrl.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("error parsing DestinationURL: %w", err)
	}

	if 255 <= len(uri.String()) {
		return nil, errors.New("value is too long! Max 255 characters allowed")
	}

	if uri.Scheme() != "http" && uri.Scheme() != "https" {
		return nil, fmt.Errorf("invalid scheme %s", uri.Scheme())
	}

	return &DestinationURL{url: uri}, nil
}
