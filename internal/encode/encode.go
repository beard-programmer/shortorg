package encode

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type UrlWasEncoded struct {
	Token TokenStandard
}

func NewEncodeFunc(
	keyIssuer KeyIssuer,
	urlParser UrlParser,
	codec Encoder,
	logger *zap.SugaredLogger,
	urlWasEncodedChan chan<- UrlWasEncoded,
) func(context.Context, EncodingRequest) (*UrlWasEncoded, error) {
	return func(ctx context.Context, r EncodingRequest) (*UrlWasEncoded, error) {
		return encode(ctx, keyIssuer, urlParser, codec, logger, urlWasEncodedChan, r)
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
	keyIssuer KeyIssuer,
	urlParser UrlParser,
	codec Encoder,
	logger *zap.SugaredLogger,
	urlWasEncodedChan chan<- UrlWasEncoded,
	request EncodingRequest,
) (*UrlWasEncoded, error) {
	validatedRequest, err := NewValidatedRequest(
		urlParser,
		request,
	)

	if err != nil {
		return nil, ValidationError{Err: err}
	}

	unclaimedKey, err := keyIssuer.Issue(ctx)
	if err != nil {
		return nil, InfrastructureError{Err: fmt.Errorf("failed to generate unclaimedKey: %w", err)}
	}

	token, err := NewToken(codec, *unclaimedKey, validatedRequest.TokenHost, validatedRequest.OriginalURL)

	if err != nil {
		return nil, ApplicationError{Err: fmt.Errorf("failed to make token: %w", err)}
	}

	event := UrlWasEncoded{*token}
	go func() {
		urlWasEncodedChan <- event
	}()

	return &event, nil
}
