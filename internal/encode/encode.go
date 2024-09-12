package encode

import (
	"context"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/simple_types"
	"go.uber.org/zap"
)

type EncodedUrl struct {
	URL   OriginalURL
	Token TokenStandard
}

func NewEncodeFunc(
	identityProvider Identities,
	urlProvider UrlProvider,
	codecProvider CodecProvider,
	logger *zap.SugaredLogger,
	encodedUrlsChan chan<- EncodedUrl,
) func(context.Context, EncodingRequest) (*EncodedUrl, error) {
	return func(ctx context.Context, r EncodingRequest) (*EncodedUrl, error) {
		return encode(ctx, identityProvider, urlProvider, codecProvider, logger, encodedUrlsChan, r)
	}
}

type ValidationError struct {
	Err error
}

func (e ValidationError) Error() string {
	return e.Err.Error()
}

type InfrastructureError struct {
	Err error
}

func (e InfrastructureError) Error() string {
	return e.Err.Error()
}

type ApplicationError struct {
	Err error
}

func (e ApplicationError) Error() string {
	return e.Err.Error()
}

func encode(
	ctx context.Context,
	identityProvider Identities,
	urlProvider UrlProvider,
	codecProvider CodecProvider,
	logger *zap.SugaredLogger,
	encodedUrlsChan chan<- EncodedUrl,
	request EncodingRequest,
) (*EncodedUrl, error) {
	validatedRequest, err := NewValidatedRequest(
		urlProvider.Parse,
		request.OriginalUrl(),
		request.Host(),
	)

	if err != nil {
		return nil, ValidationError{Err: err}
	}

	identity, err := identityProvider.GenerateOne(ctx)
	if err != nil {
		return nil, InfrastructureError{Err: fmt.Errorf("failed to generate identity: %w", err)}
	}

	token, err := NewToken(codecProvider, *identity, validatedRequest.TokenHost, validatedRequest.OriginalURL)

	if err != nil {
		return nil, ApplicationError{Err: fmt.Errorf("failed to make token: %w", err)}
	}

	encodedUrl := EncodedUrl{
		URL:   validatedRequest.OriginalURL,
		Token: *token,
	}
	go func() {
		encodedUrlsChan <- encodedUrl
	}()

	return &encodedUrl, nil
}

type Identity = simple_types.IntegerBase58Exp5To6

func NewIdentity(value int64) (*Identity, error) {
	return simple_types.NewIntegerBase58Exp5To6(value)
}
