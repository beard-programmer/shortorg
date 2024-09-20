package encode

import (
	"context"
	"fmt"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/core"
)

type URLWasEncoded struct {
	NonBrandedLink core.NonBrandedLink
}

type Fn = func(context.Context, EncodingRequest) (*URLWasEncoded, error)

func NewEncodeFn(
	tokenKeyStore LinkKeyStore,
	logger *appLogger.AppLogger,
	urlWasEncodedChan chan<- URLWasEncoded,
) Fn {
	return func(ctx context.Context, r EncodingRequest) (*URLWasEncoded, error) {
		return encode(ctx, tokenKeyStore, logger, urlWasEncodedChan, r)
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
	linkKeyStore LinkKeyStore,
	logger *appLogger.AppLogger,
	urlWasEncodedChan chan<- URLWasEncoded,
	request EncodingRequest,
) (*URLWasEncoded, error) {
	validatedRequest, err := NewValidatedRequest(
		request,
	)

	if err != nil {
		return nil, ValidationError{Err: err}
	}

	unclaimedKey, err := linkKeyStore.Issue(ctx)
	if err != nil {
		return nil, InfrastructureError{Err: fmt.Errorf("failed to generate unclaimedKey: %w", err)}
	}

	token, err := core.NewNonBrandedLink(*unclaimedKey, validatedRequest.TokenHost, validatedRequest.OriginalURL)

	if err != nil {
		return nil, ApplicationError{Err: fmt.Errorf("failed to make token: %w", err)}
	}

	event := URLWasEncoded{*token}
	go func() {
		urlWasEncodedChan <- event
	}()

	return &event, nil
}
