package encode

import (
	"context"
	"fmt"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/core"
)

type UrlWasEncoded struct {
	Token core.NonBrandedLink
}

type Fn = func(context.Context, EncodingRequest) (*UrlWasEncoded, error)

func NewEncodeFn(
	tokenKeyStore LinkKeyStore,
	urlParser URLParser,
	logger *appLogger.AppLogger,
	urlWasEncodedChan chan<- UrlWasEncoded,
) Fn {
	return func(ctx context.Context, r EncodingRequest) (*UrlWasEncoded, error) {
		return encode(ctx, tokenKeyStore, urlParser, logger, urlWasEncodedChan, r)
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
	keyIssuer LinkKeyStore,
	urlParser URLParser,
	logger *appLogger.AppLogger,
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

	token, err := core.NewNonBrandedLink(*unclaimedKey, validatedRequest.TokenHost, validatedRequest.OriginalURL)

	if err != nil {
		return nil, ApplicationError{Err: fmt.Errorf("failed to make token: %w", err)}
	}

	event := UrlWasEncoded{*token}
	go func() {
		urlWasEncodedChan <- event
	}()

	return &event, nil
}
