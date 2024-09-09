package encode

import (
	"context"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/common"
	"go.uber.org/zap"
)

type UrlWasEncoded struct {
	URL   string
	Token TokenStandard
}

type Request struct {
	URL          string
	EncodeAtHost *string
}

type IdentityProvider interface {
	ProduceTokenIdentifier(ctx context.Context) (*TokenIdentifier, error)
}

type EncodedUrl struct {
	Url             string
	TokenIdentifier *TokenIdentifier
}

func Encode(
	ctx context.Context,
	identityProvider IdentityProvider,
	parseUrl func(string) (URL, error),
	logger *zap.SugaredLogger,
	request Request,
) (*UrlWasEncoded, error) {
	validatedRequest, err := FromUnvalidatedRequest(
		parseUrl,
		request.URL,
		request.EncodeAtHost,
	)

	if err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	tokenIdentifier, err := identityProvider.ProduceTokenIdentifier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate identity: %w", err)
	}

	token, err := NewToken(&common.CodedBase58{}, *tokenIdentifier, validatedRequest.TokenHost)

	if err != nil {
		return nil, fmt.Errorf("failed to make token: %w", err)
	}

	//newCtx, cancelNewCtx := context.WithTimeout(context.Background(), 30*time.Second)
	//go func(ctx context.Context, validatedRequest *RequestValidated, tokenIdentifier *TokenIdentifier) {
	//	defer cancelNewCtx()
	//	err := saveEncodedUrlProvider.SaveEncodedURL(ctx, validatedRequest.OriginalURL, tokenIdentifier.Value())
	//	if err != nil {
	//		logger.Errorf("failed to save encoded url: %v", err)
	//	}
	//}(newCtx, validatedRequest, tokenIdentifier)

	return &UrlWasEncoded{
		URL:   validatedRequest.OriginalURL,
		Token: *token,
	}, nil
}
