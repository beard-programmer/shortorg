package decode

import (
	"fmt"

	"github.com/beard-programmer/shortorg/internal/core"
)

type decodingRequest interface {
	Url() string
}

type validatedRequest struct {
	ShortURL shortUrl
}

func newValidatedRequest(urlParser UrlParser, request decodingRequest) (*validatedRequest, error) {
	shortURL, err := newShortUrl(urlParser, request.Url())
	if err != nil {
		return nil, err
	}

	return &validatedRequest{ShortURL: *shortURL}, nil

}

type shortUrl struct {
	linkSlug core.LinkSlug
	linkHost core.LinkHost
}

func newShortUrl(urlParser UrlParser, url string) (*shortUrl, error) {
	uri, err := urlParser.Parse(url)
	if err != nil {
		return nil, err
	}

	scheme := uri.Scheme()
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("invalid ShortUrl scheme: %s", scheme)
	}
	hostname := uri.Hostname()
	tokenHost, err := core.LinkHostFromString(&hostname)
	if err != nil {
		return nil, err
	}

	encodedKey, err := core.NewLinkSlug(uri.Path()[1:])
	if err != nil {
		return nil, err
	}

	return &shortUrl{linkSlug: *encodedKey, linkHost: tokenHost}, nil

}
