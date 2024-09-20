package core

import (
	"fmt"
	"net/url"
)

const (
	minURLLen = 10
	maxURLLen = 2048
)

type URL struct {
	scheme   string
	hostname string
	path     string
}

func NewURL(urlString string) (*URL, error) {
	if len(urlString) < minURLLen || maxURLLen <= len(urlString) {
		return nil, fmt.Errorf(
			"%w NewURL: urlString %s is out of range: its len must be included in %d .. %d",
			errValidation,
			urlString,
			minURLLen,
			maxURLLen-1,
		)
	}

	parsed, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, fmt.Errorf("%w NewURL: failed to parse: %v", errValidation, err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("%w NewURL: scheme %s is not supported", errValidation, parsed.Scheme)
	}

	return &URL{scheme: parsed.Scheme, hostname: parsed.Hostname(), path: parsed.Path}, nil
}

func (u *URL) String() string {
	return fmt.Sprintf("%s://%s%s", u.scheme, u.hostname, u.path)
}

func (u *URL) Hostname() string {
	return u.hostname
}

func (u *URL) Scheme() string {
	return u.scheme
}

func (u *URL) Path() string {
	return u.path
}
