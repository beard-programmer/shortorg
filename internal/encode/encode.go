package encode

import (
	"fmt"

	"github.com/beard-programmer/shortorg/internal/common"
	"github.com/beard-programmer/shortorg/internal/encode/infrastructure"
)

type UrlWasEncoded struct {
	URL        string
	ShortHost  string
	ShortToken string
}

type Request struct {
	URL          string
	EncodeAtHost *string
}

type IdentityProvider interface {
	ProduceTokenIdentifier() (*TokenIdentifier, error)
}

func Encode(
	identityProvider IdentityProvider,
	//persist func(encodedURL EncodedUrl) error,
	request Request,
) (*UrlWasEncoded, error) {
	validatedRequest, err := FromUnvalidatedRequest(
		func(s string) (URL, error) {
			return infrastructure.ParseURLString(s)
		},
		request.URL,
		request.EncodeAtHost,
	)
	if err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	tokenIdentifier, err := identityProvider.ProduceTokenIdentifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate identity: %w", err)
	}

	token, err := NewToken(&common.CodedBase58{}, *tokenIdentifier, validatedRequest.TokenHost)

	if err != nil {
		return nil, fmt.Errorf("failed to make token: %w", err)
	}

	return &UrlWasEncoded{
		URL:        validatedRequest.OriginalURL,
		ShortHost:  validatedRequest.TokenHost.Host(),
		ShortToken: token.TokenEncoded.Value(),
	}, nil

}
