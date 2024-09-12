package decode

import (
	"fmt"

	"github.com/beard-programmer/shortorg/internal/core"
)

type DecodingRequest interface {
	Url() string
}

type ValidatedRequest struct {
	ShortUrl ShortUrl
}

func (ValidatedRequest) New(urlParser UrlParser, request DecodingRequest) (*ValidatedRequest, error) {
	shortUrl, err := ShortUrl{}.new(urlParser, request.Url())
	if err != nil {
		return nil, err
	}
	return &ValidatedRequest{ShortUrl: *shortUrl}, nil

}

type ShortUrl struct {
	KeyEncoded core.TokenKeyEncoded
	Host       core.TokenHost
}

func (ShortUrl) new(urlParser UrlParser, url string) (*ShortUrl, error) {
	uri, err := urlParser.Parse(url)
	if err != nil {
		return nil, err
	}

	scheme := uri.Scheme()
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("invalid ShortUrl scheme: %s", scheme)
	}
	hostname := uri.Hostname()
	tokenHost, err := core.TokenHostFromString(&hostname)
	if err != nil {
		return nil, err
	}

	encodedKey, err := core.TokenKeyEncoded{}.New(uri.Path()[1:])
	if err != nil {
		return nil, err
	}

	return &ShortUrl{KeyEncoded: *encodedKey, Host: tokenHost}, nil

}
