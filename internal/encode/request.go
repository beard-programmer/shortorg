package encode

import (
	"fmt"

	"github.com/beard-programmer/shortorg/internal/core"
)

type EncodingRequest interface {
	OriginalUrl() string
	Host() *string
}

type ValidatedRequest struct {
	OriginalURL core.DestinationURL
	TokenHost   core.LinkHost
}

func NewValidatedRequest(request EncodingRequest) (*ValidatedRequest, error) {
	destinationURL, err := core.NewURL(request.OriginalUrl())
	if err != nil {
		return nil, fmt.Errorf("parsing original url failed: %w", err)
	}

	linkHost, err := core.NewLinkHost(request.Host())
	if err != nil {
		return nil, err
	}

	return &ValidatedRequest{
		OriginalURL: *destinationURL,
		TokenHost:   *linkHost,
	}, nil
}
