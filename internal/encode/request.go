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

func NewValidatedRequest(urlParser UrlParser, request EncodingRequest) (*ValidatedRequest, error) {
	originalUrl, err := OriginalURLFromString(urlParser, request.OriginalUrl())
	if err != nil {
		return nil, fmt.Errorf("parsing original url failed: %w", err)
	}

	tokenHost, err := TokenHostFromString(request.Host())
	if err != nil {
		return nil, err
	}

	if originalUrl.url.Hostname() == tokenHost.Hostname() {
		return nil, fmt.Errorf("request validation failed: cannot encode self")
	}

	return &ValidatedRequest{
		OriginalURL: *originalUrl,
		TokenHost:   tokenHost,
	}, nil
}
