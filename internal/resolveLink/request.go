package resolveLink

import (
	"github.com/beard-programmer/shortorg/internal/core"
)

type resolveLinkRequest interface {
	Url() string
}

type validatedRequest struct {
	ShortURL shortUrl
}

func newValidatedRequest(request resolveLinkRequest) (*validatedRequest, error) {
	shortURL, err := newShortUrl(request.Url())
	if err != nil {
		return nil, err
	}

	return &validatedRequest{ShortURL: *shortURL}, nil

}

type shortUrl struct {
	linkSlug core.LinkSlug
	linkHost core.LinkHost
}

func newShortUrl(url string) (*shortUrl, error) {
	uri, err := core.NewURL(url)
	hostname := uri.Hostname()
	tokenHost, err := core.NewLinkHost(&hostname)
	if err != nil {
		return nil, err
	}

	encodedKey, err := core.NewLinkSlug(uri.Path()[1:])
	if err != nil {
		return nil, err
	}

	return &shortUrl{linkSlug: *encodedKey, linkHost: *tokenHost}, nil

}
